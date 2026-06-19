package teamdata

import (
	"context"
	"log"

	"wc-system/adapter/theodds"
)

// LoadLatestOddsFromDB hydrates the in-memory odds cache from historical_odds.
func (s *Store) LoadLatestOddsFromDB(ctx context.Context) error {
	if s.pool == nil {
		return nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT m.home_id, m.away_id,
		       ho.home_odds, ho.draw_odds, ho.away_odds,
		       COALESCE(ho.totals_line, 0), COALESCE(ho.over_odds, 0), COALESCE(ho.under_odds, 0)
		FROM historical_odds ho
		JOIN matches m ON m.id = ho.match_id
		WHERE ho.source = $1
		  AND m.home_id IS NOT NULL AND m.away_id IS NOT NULL
	`, theodds.SourceName())
	if err != nil {
		return err
	}
	defer rows.Close()

	n := 0
	s.mu.Lock()
	defer s.mu.Unlock()
	for rows.Next() {
		var homeID, awayID string
		var snap OddsSnapshot
		if err := rows.Scan(
			&homeID, &awayID,
			&snap.HomeOdds, &snap.DrawOdds, &snap.AwayOdds,
			&snap.TotalsLine, &snap.OverOdds, &snap.UnderOdds,
		); err != nil {
			continue
		}
		s.latestOdds[homeID+":"+awayID] = snap
		n++
	}
	log.Printf("teamdata: loaded %d odds snapshots from DB", n)
	return nil
}
