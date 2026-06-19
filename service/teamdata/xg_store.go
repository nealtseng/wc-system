package teamdata

import (
	"context"
	"fmt"
	"log"
	"time"

	"wc-system/adapter/fbref"
	"wc-system/adapter/fbrefcsv"
	"wc-system/catalog"
)

const (
	xgShrinkageK        = 5.0
	xgLowSampleThreshold = 5
)

// FBrefSyncResult summarizes one FBref-only sync pass.
type FBrefSyncResult struct {
	TeamsOK    int    `json:"teams_ok"`
	TeamsTotal int    `json:"teams_total"`
	TeamsFail  int    `json:"teams_fail"`
	Message    string `json:"message"`
}

// BootstrapXG loads persisted xG from PostgreSQL and applies optional CSV seeds.
func (s *Store) BootstrapXG(ctx context.Context) error {
	if err := s.loadXGFromDB(ctx); err != nil {
		return err
	}
	return s.applyXGSeedCSV(ctx)
}

func (s *Store) loadXGFromDB(ctx context.Context) error {
	if s.pool == nil {
		return nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, avg_xg_for, avg_xg_source, avg_xg_match_count, avg_xg_updated_at
		FROM teams
		WHERE avg_xg_for IS NOT NULL AND avg_xg_for > 0
	`)
	if err != nil {
		return fmt.Errorf("load xG from DB: %w", err)
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		var (
			teamID     string
			xg         float64
			source     *string
			matchCount int
			updatedAt  *time.Time
		)
		if err := rows.Scan(&teamID, &xg, &source, &matchCount, &updatedAt); err != nil {
			return fmt.Errorf("scan team xG: %w", err)
		}
		rec, ok := s.teams[teamID]
		if !ok {
			if cat, found := catalog.ByID(teamID); found {
				rec = Record{ID: teamID, Name: cat.Name, Group: cat.Group, ISO2: cat.ISO2}
			} else {
				rec = Record{ID: teamID}
			}
		}
		rec.AvgXGFor = xg
		rec.AvgXGMatchCount = matchCount
		if source != nil {
			rec.AvgXGSource = *source
		}
		if updatedAt != nil {
			rec.AvgXGUpdatedAt = *updatedAt
		}
		s.teams[teamID] = rec
	}
	return rows.Err()
}

func (s *Store) applyXGSeedCSV(ctx context.Context) error {
	path := s.fbrefXGCsv
	if path == "" {
		path = "data/fbref_xg.csv"
	}
	rows, err := fbrefcsv.Load(path)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	applied := 0
	for _, row := range rows {
		s.mu.RLock()
		rec, ok := s.teams[row.TeamID]
		s.mu.RUnlock()
		if ok && rec.AvgXGFor > 0 {
			continue
		}
		s.setTeamXG(ctx, row.TeamID, row.AvgXGFor, "manual", 0)
		applied++
	}
	if applied > 0 {
		log.Printf("teamdata: applied %d manual xG seeds from %s", applied, path)
	}
	return nil
}

func (s *Store) setTeamXG(ctx context.Context, teamID string, xg float64, source string, matchCount int) {
	now := time.Now().UTC()

	s.mu.Lock()
	rec, ok := s.teams[teamID]
	if !ok {
		if cat, found := catalog.ByID(teamID); found {
			rec = Record{ID: teamID, Name: cat.Name, Group: cat.Group, ISO2: cat.ISO2}
		} else {
			rec = Record{ID: teamID}
		}
	}
	rec.AvgXGFor = xg
	rec.AvgXGSource = source
	rec.AvgXGMatchCount = matchCount
	rec.AvgXGUpdatedAt = now
	s.teams[teamID] = rec
	s.mu.Unlock()

	if s.pool == nil {
		return
	}

	name := teamID
	if cat, ok := catalog.ByID(teamID); ok {
		name = cat.Name
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO teams (id, name, elo, avg_xg_for, avg_xg_updated_at, avg_xg_source, avg_xg_match_count)
		VALUES ($1, $2, 1500, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			avg_xg_for = EXCLUDED.avg_xg_for,
			avg_xg_updated_at = EXCLUDED.avg_xg_updated_at,
			avg_xg_source = EXCLUDED.avg_xg_source,
			avg_xg_match_count = EXCLUDED.avg_xg_match_count,
			updated_at = NOW()
	`, teamID, name, xg, now, source, matchCount)
	if err != nil {
		log.Printf("teamdata: persist xG %s: %v", teamID, err)
	}
}

// SyncFBref scrapes FBref for all catalog teams and keeps stale values on failure.
func (s *Store) SyncFBref(ctx context.Context) (FBrefSyncResult, error) {
	s.fbrefSyncMu.Lock()
	defer s.fbrefSyncMu.Unlock()

	teamList := catalog.All()
	start := time.Now()

	cl := fbref.NewClient()
	if err := cl.Warmup(ctx); err != nil {
		log.Printf("teamdata: fbref warmup: %v", err)
	}

	var ok, fail, total int
	for _, t := range teamList {
		if err := ctx.Err(); err != nil {
			return FBrefSyncResult{}, err
		}
		slug, hasSlug := catalog.FBrefSlug(t.ID)
		if !hasSlug {
			continue
		}
		total++

		result, err := cl.FetchTeamXG(ctx, slug)
		if err != nil {
			fail++
			log.Printf("teamdata: fbref %s: %v", t.ID, err)
			time.Sleep(fbref.DelayBetweenTeams())
			continue
		}

		s.setTeamXG(ctx, t.ID, result.Avg, "fbref", result.MatchCount)
		ok++
		time.Sleep(fbref.DelayBetweenTeams())
	}

	out := FBrefSyncResult{
		TeamsOK:    ok,
		TeamsTotal: total,
		TeamsFail:  fail,
	}
	switch {
	case total == 0:
		out.Message = "no FBref slugs configured"
	case ok == 0:
		out.Message = fmt.Sprintf("0/%d teams updated — stale DB/cache values kept", total)
	case fail > 0:
		out.Message = fmt.Sprintf("%d/%d teams updated (%d failed, stale cache kept)", ok, total, fail)
	default:
		out.Message = fmt.Sprintf("%d/%d teams updated", ok, total)
	}

	state := ScraperState{Name: scraperFBref, LastFetch: start}
	switch {
	case ok == 0 && total > 0:
		state.Status = "offline"
		state.Message = out.Message
	case fail > 0:
		state.Status = "degraded"
		state.Message = out.Message
	default:
		state.Status = "ok"
	}
	s.mu.Lock()
	s.scrapers[scraperFBref] = state
	s.mu.Unlock()

	if ok == 0 && total > 0 {
		return out, fmt.Errorf(out.Message)
	}
	return out, nil
}

func (s *Store) expectedGoals(id string) (float64, string) {
	rec, ok := s.Get(id)
	if !ok {
		return 1.2, "default"
	}

	prior := rec.AvgGoalsFor
	if prior <= 0 {
		prior = 1.2
	}

	if rec.AvgXGFor > 0 {
		source := rec.AvgXGSource
		if source == "" {
			source = "fbref"
		}
		if source == "manual" {
			return rec.AvgXGFor, "manual"
		}

		n := rec.AvgXGMatchCount
		if n >= xgLowSampleThreshold {
			return rec.AvgXGFor, source
		}

		shrunk := (float64(n)*rec.AvgXGFor + xgShrinkageK*prior) / (float64(n) + xgShrinkageK)
		return shrunk, source + "_shrunk"
	}

	if rec.AvgGoalsFor > 0 {
		return rec.AvgGoalsFor, "kaggle"
	}
	return 1.2, "default"
}

// XGSourceFor reports which layer supplied Poisson xG for one team.
func (s *Store) XGSourceFor(id string) string {
	_, src := s.expectedGoals(id)
	return src
}
