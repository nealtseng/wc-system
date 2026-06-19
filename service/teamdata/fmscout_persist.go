package teamdata

import (
	"context"
	"encoding/json"
	"log"
	"sort"

	"wc-system/adapter/fmscout"
)

func computeFMScoutStrength(players []fmscout.PlayerAttribute) float64 {
	if len(players) == 0 {
		return 0
	}
	sorted := append([]fmscout.PlayerAttribute(nil), players...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Overall > sorted[j].Overall
	})
	n := min(11, len(sorted))
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += sorted[i].Overall
	}
	return sum / (float64(n) * 100.0)
}

func (s *Store) persistFMScoutPlayers(ctx context.Context, teamID string, players []fmscout.PlayerAttribute) {
	if s.pool == nil {
		return
	}
	for _, p := range players {
		attrsJSON, err := json.Marshal(p.Attributes)
		if err != nil {
			continue
		}
		_, err = s.pool.Exec(ctx, `
			INSERT INTO player_stats (team_id, name, position, overall, source, attributes)
			VALUES ($1, $2, $3, $4, 'fmscout', $5::jsonb)
			ON CONFLICT (team_id, name) DO UPDATE SET
				position = EXCLUDED.position,
				overall = EXCLUDED.overall,
				source = EXCLUDED.source,
				attributes = EXCLUDED.attributes,
				imported_at = NOW()
		`, teamID, p.Name, p.Position, p.Overall, string(attrsJSON))
		if err != nil {
			log.Printf("teamdata: fmscout persist %s/%s: %v", teamID, p.Name, err)
		}
	}
}
