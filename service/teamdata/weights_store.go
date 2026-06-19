package teamdata

import (
	"context"
	"fmt"
	"sort"
)

// Weights returns a copy of the current model weights.
func (s *Store) Weights() ModelWeights {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.weights
}

// LoadWeights reads model_config from the database into the store.
func (s *Store) LoadWeights(ctx context.Context) error {
	if s.pool == nil {
		return nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT key, value FROM model_config
		WHERE key IN ('w1','w2','w3','clip_delta','delta_max','kelly_scale')
	`)
	if err != nil {
		return fmt.Errorf("teamdata: load weights: %w", err)
	}
	defer rows.Close()

	w := defaultWeights()
	for rows.Next() {
		var key string
		var val float64
		if err := rows.Scan(&key, &val); err != nil {
			continue
		}
		switch key {
		case "w1":
			w.W1 = val
		case "w2":
			w.W2 = val
		case "w3":
			w.W3 = val
		case "clip_delta":
			w.ClipDelta = val
		case "delta_max":
			w.DeltaMax = val
		case "kelly_scale":
			w.KellyScale = val
		}
	}

	s.mu.Lock()
	s.weights = w
	s.mu.Unlock()
	return nil
}

// SaveWeights upserts model weights to the database and updates the in-memory copy.
func (s *Store) SaveWeights(ctx context.Context, w ModelWeights) error {
	if s.pool == nil {
		s.mu.Lock()
		s.weights = w
		s.mu.Unlock()
		return nil
	}

	pairs := []struct {
		key string
		val float64
	}{
		{"w1", w.W1},
		{"w2", w.W2},
		{"w3", w.W3},
		{"clip_delta", w.ClipDelta},
		{"delta_max", w.DeltaMax},
		{"kelly_scale", w.KellyScale},
	}
	for _, p := range pairs {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO model_config (key, value) VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
		`, p.key, p.val)
		if err != nil {
			return fmt.Errorf("teamdata: save weight %s: %w", p.key, err)
		}
	}

	s.mu.Lock()
	s.weights = w
	s.mu.Unlock()
	return nil
}

// LatestOdds returns the in-memory odds snapshot for a fixture.
func (s *Store) LatestOdds(homeID, awayID string) (OddsSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.latestOdds[homeID+":"+awayID]
	return snap, ok
}

// SquadStrength returns normalized squad strength from SoFIFA prof ratings [0,1].
func (s *Store) SquadStrength(teamID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return squadStrengthFromPlayers(s.squads[teamID])
}

func squadStrengthFromPlayers(players []SquadPlayer) float64 {
	if len(players) == 0 {
		return 0.5
	}
	sorted := append([]SquadPlayer(nil), players...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Prof > sorted[j].Prof
	})
	n := min(11, len(sorted))
	sum := 0
	for i := 0; i < n; i++ {
		sum += sorted[i].Prof
	}
	return float64(sum) / (float64(n) * 20.0)
}
