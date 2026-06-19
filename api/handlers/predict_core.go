package handlers

import (
	"time"

	"wc-system/engine/poisson"
	"wc-system/service/predict"
	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
)

type scoreEntry struct {
	Home int     `json:"home"`
	Away int     `json:"away"`
	Prob float64 `json:"prob"`
}

type modelOutput = predict.Output

func computeModel(
	store *teamdata.Store,
	homeID, awayID string,
	iterations int,
	seed int64,
	isNeutral bool,
	overrideWeights *teamdata.ModelWeights,
) (modelOutput, error) {
	return predict.Compute(store, homeID, awayID, iterations, seed, isNeutral, overrideWeights)
}

var errTeamNotSynced = predict.ErrTeamNotSynced

func topScoresJSON(p poisson.PoissonResult) []scoreEntry {
	out := make([]scoreEntry, len(p.TopScores))
	for i, s := range p.TopScores {
		out[i] = scoreEntry{Home: s.Home, Away: s.Away, Prob: s.Prob}
	}
	return out
}

func monteCarloSeed() int64 {
	return time.Now().UnixNano()
}

// wcNeutralDefault returns true when the client omits the neutral flag.
// World Cup 2026 fixtures in USA/Mexico/Canada are treated as neutral for both sides.
func wcNeutralDefault(neutral *bool) bool {
	if neutral == nil {
		return true
	}
	return *neutral
}

func predictResponseJSON(homeID, awayID string, out predict.Output, xgSource string) gin.H {
	resp := gin.H{
		"home_team_id": homeID,
		"away_team_id": awayID,
		"neutral_venue": out.IsNeutral,
		"elo": gin.H{
			"home": out.EloHome,
			"away": out.EloAway,
		},
		"poisson": gin.H{
			"home_lambda": out.Poisson.LambdaHome,
			"away_lambda": out.Poisson.LambdaAway,
			"home_win":    out.Poisson.HomeWin,
			"draw":        out.Poisson.Draw,
			"away_win":    out.Poisson.AwayWin,
			"top_scores":  topScoresJSON(out.Poisson),
		},
		"poisson_w2": gin.H{
			"home_win": out.PoissonW2Win.Home,
			"draw":     out.PoissonW2Win.Draw,
			"away_win": out.PoissonW2Win.Away,
		},
		"lambda_layers": gin.H{
			"w2": gin.H{"home": out.LambdaW2.Home, "away": out.LambdaW2.Away},
			"w1": gin.H{"home": out.LambdaW1.Home, "away": out.LambdaW1.Away},
			"w3": gin.H{
				"home":          out.LambdaW3.Home,
				"away":          out.LambdaW3.Away,
				"implied_total": out.ImpliedTotalW3,
				"total_source":  out.W3TotalSource,
			},
		},
		"lambda_blend": gin.H{
			"home": out.LambdaBlend.Home,
			"away": out.LambdaBlend.Away,
		},
		"lambda_contrib": gin.H{
			"w2": gin.H{"home": out.LambdaContribW2.Home, "away": out.LambdaContribW2.Away},
			"w1": gin.H{"home": out.LambdaContribW1.Home, "away": out.LambdaContribW1.Away},
			"w3": gin.H{"home": out.LambdaContribW3.Home, "away": out.LambdaContribW3.Away},
		},
		"weights": out.Weights,
		"p_final": out.PFinal,
		"w3_implied": gin.H{
			"home": out.W3Implied.Home,
			"draw": out.W3Implied.Draw,
			"away": out.W3Implied.Away,
		},
		"blend": gin.H{
			"w2": gin.H{
				"home": out.BlendW2.Home,
				"draw": out.BlendW2.Draw,
				"away": out.BlendW2.Away,
			},
			"w3": gin.H{
				"home": out.BlendW3.Home,
				"draw": out.BlendW3.Draw,
				"away": out.BlendW3.Away,
			},
			"w1_delta": out.ClippedW1,
		},
		"signals": gin.H{
			"poisson_favors":    out.PoissonFavors,
			"final_favors":      out.FinalFavors,
			"w2_lambda_favors":  out.LambdaFavorW2,
			"w3_lambda_favors":  out.LambdaFavorW3,
			"conflict":          out.SignalConflict,
		},
		"sources": gin.H{
			"elo":      "kaggle",
			"xg":       xgSource,
			"gdp":      "worldbank",
			"wiki":     "wikimedia",
			"w1_delta": out.W1MicroDelta,
			"w3":       out.W3Source,
		},
	}
	if out.W3Odds != nil {
		resp["w3_odds"] = gin.H{
			"home": out.W3Odds.Home,
			"draw": out.W3Odds.Draw,
			"away": out.W3Odds.Away,
		}
	}
	if out.W3TotalsOdds != nil {
		resp["w3_totals"] = gin.H{
			"line":  out.W3TotalsOdds.Line,
			"over":  out.W3TotalsOdds.Over,
			"under": out.W3TotalsOdds.Under,
		}
	}
	return resp
}
