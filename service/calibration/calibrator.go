package calibration

import (
	"context"
	"errors"
	"fmt"

	"wc-system/service/predict"
	"wc-system/service/teamdata"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CalibrationResult holds the best weight combination from grid search.
type CalibrationResult struct {
	W1, W2, W3 float64
	BrierScore  float64
	MatchesUsed int
}

var errInsufficientData = errors.New("insufficient data")

type matchData struct {
	homeID, awayID           string
	actualHome, actualDraw, actualAway float64
}

// Calibrate runs grid search over w1/w2/w3 at 5% steps using finished WC matches.
func Calibrate(ctx context.Context, pool *pgxpool.Pool, store *teamdata.Store) (CalibrationResult, error) {
	if pool == nil {
		return CalibrationResult{}, errInsufficientData
	}

	rows, err := pool.Query(ctx, `
		SELECT home_id, away_id, home_score, away_score
		FROM matches
		WHERE finished = true
		  AND home_id IS NOT NULL AND away_id IS NOT NULL
		  AND home_score IS NOT NULL AND away_score IS NOT NULL
	`)
	if err != nil {
		return CalibrationResult{}, fmt.Errorf("calibration: query matches: %w", err)
	}
	defer rows.Close()

	var matches []matchData
	for rows.Next() {
		var homeID, awayID string
		var homeScore, awayScore int
		if err := rows.Scan(&homeID, &awayID, &homeScore, &awayScore); err != nil {
			continue
		}
		m := matchData{homeID: homeID, awayID: awayID}
		switch {
		case homeScore > awayScore:
			m.actualHome, m.actualDraw, m.actualAway = 1, 0, 0
		case homeScore < awayScore:
			m.actualHome, m.actualDraw, m.actualAway = 0, 0, 1
		default:
			m.actualHome, m.actualDraw, m.actualAway = 0, 1, 0
		}
		matches = append(matches, m)
	}
	if len(matches) < 3 {
		return CalibrationResult{}, errInsufficientData
	}

	base := store.Weights()
	bestBS := 1e9
	var bestW1, bestW2, bestW3 float64

	for w1 := 0.10; w1 <= 0.50+1e-9; w1 += 0.05 {
		for w2 := 0.10; w2 <= 0.50+1e-9; w2 += 0.05 {
			w3 := 1.0 - w1 - w2
			if w3 < 0.10-1e-9 || w3 > 0.50+1e-9 {
				continue
			}
			testWeights := teamdata.ModelWeights{
				W1: w1, W2: w2, W3: w3,
				ClipDelta: base.ClipDelta, DeltaMax: base.DeltaMax, KellyScale: base.KellyScale,
			}
			totalBS := 0.0
			used := 0
			for _, m := range matches {
				pFinal, err := predict.Compute(store, m.homeID, m.awayID, 5000, 42, true, &testWeights)
				if err != nil {
					continue
				}
				totalBS += brierScore(
					pFinal.PFinal["home"], pFinal.PFinal["draw"], pFinal.PFinal["away"],
					m.actualHome, m.actualDraw, m.actualAway,
				)
				used++
			}
			if used == 0 {
				continue
			}
			meanBS := totalBS / float64(used)
			if meanBS < bestBS {
				bestBS = meanBS
				bestW1, bestW2, bestW3 = w1, w2, w3
			}
		}
	}

	return CalibrationResult{
		W1: bestW1, W2: bestW2, W3: bestW3,
		BrierScore:  bestBS,
		MatchesUsed: len(matches),
	}, nil
}

func brierScore(pH, pD, pA, aH, aD, aA float64) float64 {
	return (pH-aH)*(pH-aH) + (pD-aD)*(pD-aD) + (pA-aA)*(pA-aA)
}
