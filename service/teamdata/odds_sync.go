package teamdata

import (
	"context"
	"log"
	"time"

	"wc-system/adapter/theodds"
)

func scoreArg(score *int) any {
	if score == nil {
		return nil
	}
	return *score
}

type dbMatchRow struct {
	ID       int64
	HomeID   string
	AwayID   string
	HomeName string
	AwayName string
	Kickoff  time.Time
}

func (s *Store) syncTheOdds(ctx context.Context) error {
	events, err := theodds.FetchWC2026(ctx, s.oddsAPIKey)
	if err != nil {
		return err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, COALESCE(home_id,''), COALESCE(away_id,''),
		       COALESCE(home_name,''), COALESCE(away_name,''), kickoff
		FROM matches
		WHERE home_id IS NOT NULL AND away_id IS NOT NULL
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var fixtures []dbMatchRow
	for rows.Next() {
		var m dbMatchRow
		if err := rows.Scan(&m.ID, &m.HomeID, &m.AwayID, &m.HomeName, &m.AwayName, &m.Kickoff); err != nil {
			continue
		}
		fixtures = append(fixtures, m)
	}

	matched := 0
	now := time.Now().UTC()
	log.Printf("teamdata: odds sync fetched %d events from The Odds API", len(events))
	for _, ev := range events {
		matchID, homeID, awayID := findOddsMatchWithTeams(fixtures, ev)
		if matchID == 0 {
			continue
		}
		_, err := s.pool.Exec(ctx, `
			INSERT INTO historical_odds (
				match_id, source, recorded_at,
				home_odds, draw_odds, away_odds,
				totals_line, over_odds, under_odds
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (match_id, source) DO UPDATE SET
				recorded_at = EXCLUDED.recorded_at,
				home_odds = EXCLUDED.home_odds,
				draw_odds = EXCLUDED.draw_odds,
				away_odds = EXCLUDED.away_odds,
				totals_line = EXCLUDED.totals_line,
				over_odds = EXCLUDED.over_odds,
				under_odds = EXCLUDED.under_odds
		`, matchID, theodds.SourceName(), now,
			ev.HomeOdds, ev.DrawOdds, ev.AwayOdds,
			nullFloat(ev.TotalsLine), nullFloat(ev.OverOdds), nullFloat(ev.UnderOdds))
		if err != nil {
			log.Printf("teamdata: odds upsert match %d: %v", matchID, err)
			continue
		}
		if homeID != "" && awayID != "" {
			key := homeID + ":" + awayID
			snap := OddsSnapshot{
				HomeOdds:   ev.HomeOdds,
				DrawOdds:   ev.DrawOdds,
				AwayOdds:   ev.AwayOdds,
				TotalsLine: ev.TotalsLine,
				OverOdds:   ev.OverOdds,
				UnderOdds:  ev.UnderOdds,
			}
			s.mu.Lock()
			s.latestOdds[key] = snap
			s.mu.Unlock()
		}
		matched++
	}

	log.Printf("teamdata: odds sync matched %d/%d events to fixtures", matched, len(events))
	return nil
}

func findOddsMatchWithTeams(fixtures []dbMatchRow, ev theodds.MatchOdds) (matchID int64, homeID, awayID string) {
	const strictWindow = 3 * time.Hour
	const looseWindow = 24 * time.Hour

	var teamMatches []dbMatchRow
	for _, f := range fixtures {
		if !theodds.TeamsMatch(f.HomeName, ev.HomeTeam) ||
			!theodds.TeamsMatch(f.AwayName, ev.AwayTeam) {
			continue
		}
		teamMatches = append(teamMatches, f)
		if kickoffDiff(f.Kickoff, ev.CommenceTime) <= strictWindow {
			return f.ID, f.HomeID, f.AwayID
		}
	}
	if len(teamMatches) == 1 {
		return teamMatches[0].ID, teamMatches[0].HomeID, teamMatches[0].AwayID
	}
	if len(teamMatches) == 0 {
		return 0, "", ""
	}

	var best dbMatchRow
	bestDiff := looseWindow + 1
	for _, f := range teamMatches {
		diff := kickoffDiff(f.Kickoff, ev.CommenceTime)
		if diff <= looseWindow && diff < bestDiff {
			bestDiff = diff
			best = f
		}
	}
	if best.ID == 0 {
		return 0, "", ""
	}
	return best.ID, best.HomeID, best.AwayID
}

func findOddsMatch(fixtures []dbMatchRow, ev theodds.MatchOdds) int64 {
	id, _, _ := findOddsMatchWithTeams(fixtures, ev)
	return id
}

func kickoffDiff(a, b time.Time) time.Duration {
	diff := a.Sub(b)
	if diff < 0 {
		diff = -diff
	}
	return diff
}

func nullFloat(v float64) any {
	if v <= 0 {
		return nil
	}
	return v
}
