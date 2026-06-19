package teamdata

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// OddsScheduleEntry is one upcoming or recent odds polling window for the UI.
type OddsScheduleEntry struct {
	Label      string
	WindowKey  string
	NextAt     time.Time
	MatchLabel string
	WCMatchID  string
	Synced     bool
}

type oddsWindowDef struct {
	Label  string
	Key    string
	Before time.Duration
}

var oddsWindowDefs = []oddsWindowDef{
	{Label: "賽前 12 小時", Key: "12h", Before: 12 * time.Hour},
	{Label: "賽前 2 小時", Key: "2h", Before: 2 * time.Hour},
	{Label: "賽前 15 分鐘", Key: "15m", Before: 15 * time.Minute},
}

const (
	oddsSchedulerTick   = 1 * time.Minute
	oddsSchedulerGrace  = 10 * time.Minute
	oddsScheduleHorizon = 14 * 24 * time.Hour
)

type upcomingMatch struct {
	ID        int64
	WCMatchID string
	HomeName  string
	AwayName  string
	Kickoff   time.Time
}

// RunOddsScheduler polls for due pre-match odds windows until ctx is cancelled.
func (s *Store) RunOddsScheduler(ctx context.Context) {
	if strings.TrimSpace(s.oddsAPIKey) == "" || s.pool == nil {
		log.Println("teamdata: odds scheduler disabled (no API key or DB)")
		return
	}

	log.Println("teamdata: odds scheduler started (12h / 2h / 15m before kickoff)")
	ticker := time.NewTicker(oddsSchedulerTick)
	defer ticker.Stop()

	for {
		if err := s.runOddsSchedulerTick(ctx); err != nil {
			log.Printf("teamdata: odds scheduler tick: %v", err)
		}

		select {
		case <-ctx.Done():
			log.Println("teamdata: odds scheduler stopped")
			return
		case <-ticker.C:
		}
	}
}

func (s *Store) runOddsSchedulerTick(ctx context.Context) error {
	if strings.TrimSpace(s.oddsAPIKey) == "" || s.pool == nil {
		return nil
	}

	now := time.Now().UTC()
	matches, err := s.loadUpcomingMatches(ctx, now)
	if err != nil {
		return err
	}

	type dueEntry struct {
		MatchID   int64
		WindowKey string
	}
	var due []dueEntry

	for _, m := range matches {
		for _, w := range oddsWindowDefs {
			trigger := m.Kickoff.Add(-w.Before)
			if now.Before(trigger) || now.After(trigger.Add(oddsSchedulerGrace)) {
				continue
			}
			synced, err := s.oddsWindowSynced(ctx, m.ID, w.Key)
			if err != nil {
				return err
			}
			if synced {
				continue
			}
			due = append(due, dueEntry{MatchID: m.ID, WindowKey: w.Key})
		}
	}

	if len(due) == 0 {
		return nil
	}

	log.Printf("teamdata: odds scheduler firing for %d window(s)", len(due))
	if err := s.SyncOdds(ctx); err != nil {
		return err
	}

	for _, d := range due {
		if err := s.markOddsWindowSynced(ctx, d.MatchID, d.WindowKey, now); err != nil {
			log.Printf("teamdata: odds window mark %d/%s: %v", d.MatchID, d.WindowKey, err)
		}
	}
	return nil
}

// NextOddsSchedules returns upcoming odds polling windows derived from fixture kickoffs.
func (s *Store) NextOddsSchedules(ctx context.Context, limit int) ([]OddsScheduleEntry, error) {
	if s.pool == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 6
	}

	now := time.Now().UTC()
	matches, err := s.loadUpcomingMatches(ctx, now)
	if err != nil {
		return nil, err
	}

	synced, err := s.loadSyncedWindows(ctx)
	if err != nil {
		return nil, err
	}

	var out []OddsScheduleEntry
	for _, m := range matches {
		label := fmt.Sprintf("%s vs %s", m.HomeName, m.AwayName)
		for _, w := range oddsWindowDefs {
			trigger := m.Kickoff.Add(-w.Before)
			key := fmt.Sprintf("%d:%s", m.ID, w.Key)
			entry := OddsScheduleEntry{
				Label:      w.Label,
				WindowKey:  w.Key,
				NextAt:     trigger,
				MatchLabel: label,
				WCMatchID:  m.WCMatchID,
				Synced:     synced[key],
			}
			if entry.Synced || trigger.Before(now) {
				continue
			}
			out = append(out, entry)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].NextAt.Before(out[j].NextAt)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// SyncOdds fetches The Odds API and upserts historical_odds (no full pipeline sync).
func (s *Store) SyncOdds(ctx context.Context) error {
	var syncErr error
	s.syncScraper(scraperTheOdds, func() error {
		if strings.TrimSpace(s.oddsAPIKey) == "" {
			syncErr = fmt.Errorf("THE_ODDS_API_KEY not set")
			return syncErr
		}
		if s.pool == nil {
			return nil
		}
		syncErr = s.syncTheOdds(ctx)
		return syncErr
	})

	st := s.scrapers[scraperTheOdds]
	if st.Status == "offline" && st.Message != "" {
		return fmt.Errorf("%s", st.Message)
	}
	return syncErr
}

func (s *Store) loadUpcomingMatches(ctx context.Context, now time.Time) ([]upcomingMatch, error) {
	horizon := now.Add(oddsScheduleHorizon)
	rows, err := s.pool.Query(ctx, `
		SELECT id, COALESCE(wc_match_id,''), COALESCE(home_name,''), COALESCE(away_name,''), kickoff
		FROM matches
		WHERE home_id IS NOT NULL AND away_id IS NOT NULL
		  AND kickoff > $1
		  AND kickoff <= $2
		  AND COALESCE(finished, false) = false
		ORDER BY kickoff ASC
	`, now.Add(-2*time.Hour), horizon)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []upcomingMatch
	for rows.Next() {
		var m upcomingMatch
		if err := rows.Scan(&m.ID, &m.WCMatchID, &m.HomeName, &m.AwayName, &m.Kickoff); err != nil {
			continue
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) oddsWindowSynced(ctx context.Context, matchID int64, windowKey string) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM odds_sync_windows WHERE match_id = $1 AND window_key = $2
	`, matchID, windowKey).Scan(&n)
	return n > 0, err
}

func (s *Store) markOddsWindowSynced(ctx context.Context, matchID int64, windowKey string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO odds_sync_windows (match_id, window_key, synced_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (match_id, window_key) DO UPDATE SET synced_at = EXCLUDED.synced_at
	`, matchID, windowKey, at)
	return err
}

func (s *Store) loadSyncedWindows(ctx context.Context) (map[string]bool, error) {
	rows, err := s.pool.Query(ctx, `SELECT match_id, window_key FROM odds_sync_windows`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var matchID int64
		var key string
		if err := rows.Scan(&matchID, &key); err != nil {
			continue
		}
		out[fmt.Sprintf("%d:%s", matchID, key)] = true
	}
	return out, rows.Err()
}
