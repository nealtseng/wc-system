package predict

import (
	"math"

	"wc-system/engine/elo"
	"wc-system/engine/poisson"
	"wc-system/service/teamdata"
)

// ProbTriple holds home / draw / away probabilities or contributions.
type ProbTriple struct {
	Home float64 `json:"home"`
	Draw float64 `json:"draw"`
	Away float64 `json:"away"`
}

// OddsTriple holds decimal match odds when W3 comes from the market.
type OddsTriple struct {
	Home float64 `json:"home"`
	Draw float64 `json:"draw"`
	Away float64 `json:"away"`
}

// TotalsTriple holds O/U market odds when available.
type TotalsTriple struct {
	Line  float64 `json:"line"`
	Over  float64 `json:"over"`
	Under float64 `json:"under"`
}

// Output holds the full prediction model result.
type Output struct {
	EloHome         float64
	EloAway         float64
	Poisson         poisson.PoissonResult
	Weights         map[string]float64
	PFinal          map[string]float64
	W1MicroDelta    float64
	ClippedW1       float64
	W3Source        string
	IsNeutral       bool
	W3Implied       ProbTriple
	W3Odds          *OddsTriple
	LambdaW2        LambdaPair
	LambdaW1        LambdaPair
	LambdaW3        LambdaPair
	LambdaBlend     LambdaPair
	LambdaContribW2 LambdaPair
	LambdaContribW1 LambdaPair
	LambdaContribW3 LambdaPair
	ImpliedTotalW3  float64
	W3TotalSource   string
	W3TotalsOdds    *TotalsTriple
	BlendW2         ProbTriple
	BlendW3         ProbTriple
	LambdaFavorW2   string
	LambdaFavorW3   string
	PoissonFavors   string
	FinalFavors     string
	PoissonW2Win    ProbTriple
	SignalConflict  bool
}

var ErrTeamNotSynced = teamNotSyncedError{}

type teamNotSyncedError struct{}

func (teamNotSyncedError) Error() string {
	return "team data not synced yet — call POST /api/pipeline/sync"
}

// Compute runs the blended W1+W2+W3 lambda Poisson model.
// λ_final = W2·λ_xG/ELO + W1·λ_squad + W3·λ_market (1X2 implied total & split).
// p_final equals the Poisson outcome probabilities from λ_final (no double-counting).
func Compute(
	store *teamdata.Store,
	homeID, awayID string,
	iterations int,
	seed int64,
	isNeutral bool,
	overrideWeights *teamdata.ModelWeights,
) (Output, error) {
	homeRec, okH := store.Get(homeID)
	awayRec, okA := store.Get(awayID)
	if !okH || !okA {
		return Output{}, ErrTeamNotSynced
	}

	w := store.Weights()
	if overrideWeights != nil {
		w = *overrideWeights
	}

	ratingHome := homeRec.ELO
	ratingAway := awayRec.ELO
	eloResult := elo.Calculate(elo.ELOInput{
		TeamA:     homeID,
		TeamB:     awayID,
		RatingA:   ratingHome,
		RatingB:   ratingAway,
		IsNeutral: isNeutral,
	})

	xgHome := store.AvgGoalsFor(homeID)
	xgAway := store.AvgGoalsFor(awayID)
	lambdaW2 := lambdaFromW2(xgHome, xgAway, ratingHome, ratingAway, isNeutral)
	lambdaW1 := lambdaFromW1(lambdaW2, homeRec.PlayerStrength, awayRec.PlayerStrength)

	var w3Home, w3Draw, w3Away float64
	w3Source := "elo_fallback"
	w3TotalSource := "draw_proxy"
	var w3Odds *OddsTriple
	var w3Totals *TotalsTriple
	if odds, ok := store.LatestOdds(homeID, awayID); ok {
		w3Home, w3Draw, w3Away = removeVig(odds.HomeOdds, odds.DrawOdds, odds.AwayOdds)
		w3Source = "odds"
		w3Odds = &OddsTriple{Home: odds.HomeOdds, Draw: odds.DrawOdds, Away: odds.AwayOdds}
		if odds.HasTotals() {
			w3TotalSource = "totals"
			w3Totals = &TotalsTriple{Line: odds.TotalsLine, Over: odds.OverOdds, Under: odds.UnderOdds}
		}
	} else {
		w3Home = eloResult.ExpectedHomeWin
		w3Draw = eloResult.ExpectedDraw
		w3Away = eloResult.ExpectedAwayWin
	}

	var lambdaW3 LambdaPair
	var impliedTotal float64
	if w3Totals != nil {
		lambdaW3, impliedTotal = lambdaFromW3WithTotals(
			w3Home, w3Draw, w3Away,
			w3Totals.Over, w3Totals.Under, w3Totals.Line,
		)
	} else {
		lambdaW3, impliedTotal = lambdaFromW3(w3Home, w3Draw, w3Away)
	}
	lambdaBlend := blendLambda(w, lambdaW2, lambdaW1, lambdaW3)

	poissonResult := poisson.Simulate(poisson.PoissonInput{
		LambdaHome: lambdaBlend.Home,
		LambdaAway: lambdaBlend.Away,
		Iterations: iterations,
		Seed:       seed,
	})
	poissonW2 := poisson.Simulate(poisson.PoissonInput{
		LambdaHome: lambdaW2.Home,
		LambdaAway: lambdaW2.Away,
		Iterations: iterations,
		Seed:       seed + 997,
	})

	w1MicroDelta := store.W1MicroDelta(homeID, awayID)
	clippedW1 := math.Max(-w.DeltaMax, math.Min(w.DeltaMax, w1MicroDelta))

	pHome := poissonResult.HomeWin
	pDraw := poissonResult.Draw
	pAway := poissonResult.AwayWin

	poissonFavors := outcomeFavor(poissonResult.HomeWin, poissonResult.AwayWin)
	finalFavors := outcomeFavor(pHome, pAway)

	return Output{
		EloHome:         ratingHome,
		EloAway:         ratingAway,
		Poisson:         poissonResult,
		Weights:         map[string]float64{"w1": w.W1, "w2": w.W2, "w3": w.W3, "clip": w.ClipDelta},
		PFinal:          map[string]float64{"home": pHome, "draw": pDraw, "away": pAway},
		W1MicroDelta:    w1MicroDelta,
		ClippedW1:       clippedW1,
		W3Source:        w3Source,
		IsNeutral:       isNeutral,
		W3Implied:       ProbTriple{Home: w3Home, Draw: w3Draw, Away: w3Away},
		W3Odds:          w3Odds,
		LambdaW2:        lambdaW2,
		LambdaW1:        lambdaW1,
		LambdaW3:        lambdaW3,
		LambdaBlend:     lambdaBlend,
		LambdaContribW2: lambdaContrib(w.W2, lambdaW2),
		LambdaContribW1: lambdaContrib(w.W1, lambdaW1),
		LambdaContribW3: lambdaContrib(w.W3, lambdaW3),
		ImpliedTotalW3:  impliedTotal,
		W3TotalSource:   w3TotalSource,
		W3TotalsOdds:    w3Totals,
		BlendW2: ProbTriple{
			Home: poissonResult.HomeWin * w.W2,
			Draw: poissonResult.Draw * w.W2,
			Away: poissonResult.AwayWin * w.W2,
		},
		BlendW3: ProbTriple{
			Home: w3Home * w.W3,
			Draw: w3Draw * w.W3,
			Away: w3Away * w.W3,
		},
		PoissonFavors:   poissonFavors,
		FinalFavors:     finalFavors,
		PoissonW2Win: ProbTriple{
			Home: poissonW2.HomeWin,
			Draw: poissonW2.Draw,
			Away: poissonW2.AwayWin,
		},
		LambdaFavorW2:   lambdaFavor(lambdaW2),
		LambdaFavorW3:   lambdaFavor(lambdaW3),
		SignalConflict: lambdaFavor(lambdaW2) != lambdaFavor(lambdaW3) &&
			lambdaFavor(lambdaW2) != "even" && lambdaFavor(lambdaW3) != "even",
	}, nil
}

func outcomeFavor(home, away float64) string {
	const eps = 0.02
	switch {
	case home > away+eps:
		return "home"
	case away > home+eps:
		return "away"
	default:
		return "even"
	}
}

func removeVig(homeOdds, drawOdds, awayOdds float64) (home, draw, away float64) {
	if homeOdds <= 0 || drawOdds <= 0 || awayOdds <= 0 {
		return 0, 0, 0
	}
	ih := 1 / homeOdds
	id := 1 / drawOdds
	ia := 1 / awayOdds
	total := ih + id + ia
	if total == 0 {
		return 0, 0, 0
	}
	return ih / total, id / total, ia / total
}
