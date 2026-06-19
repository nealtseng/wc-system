package predict

import (
	"math"

	"wc-system/engine/poisson"
	"wc-system/service/teamdata"
)

// LambdaPair holds expected-goals rates for home and away before Poisson simulation.
type LambdaPair struct {
	Home float64
	Away float64
}

const (
	w1LambdaSkew       = 0.30 // squad-strength gap → λ skew
	w3DrawTotalSlope   = 1.15 // draw-prox fallback only
	w3ImpliedTotalBase = 2.55
	w3ImpliedTotalMin  = 1.75
	w3ImpliedTotalMax  = 3.25
)

func lambdaFromW2(xgHome, xgAway, eloHome, eloAway float64, isNeutral bool) LambdaPair {
	h, a := poisson.ComputeLambdas(xgHome, xgAway, eloHome, eloAway, isNeutral)
	return LambdaPair{Home: h, Away: a}
}

// lambdaFromW1 shifts scoring rates from FM / SoFIFA squad strength (PlayerStrength ∈ [0,1]).
func lambdaFromW1(base LambdaPair, strengthHome, strengthAway float64) LambdaPair {
	if strengthHome <= 0 {
		strengthHome = 0.5
	}
	if strengthAway <= 0 {
		strengthAway = 0.5
	}
	avg := (base.Home + base.Away) / 2
	if avg <= 0 {
		avg = 1.2
	}
	delta := strengthHome - strengthAway
	return LambdaPair{
		Home: avg * (1 + w1LambdaSkew*delta),
		Away: avg * (1 - w1LambdaSkew*delta),
	}
}

// lambdaFromW3 derives expected total goals and split from 1X2 implied probabilities.
// When O/U odds are unavailable, draw mass proxies market expectation for a tight / low-scoring game.
func lambdaFromW3(w3Home, w3Draw, w3Away float64) (LambdaPair, float64) {
	impliedTotal := w3ImpliedTotalBase - w3DrawTotalSlope*w3Draw
	impliedTotal = math.Max(w3ImpliedTotalMin, math.Min(w3ImpliedTotalMax, impliedTotal))
	return splitLambdaBy1X2(w3Home, w3Away, impliedTotal), impliedTotal
}

// lambdaFromW3WithTotals uses de-vigged O/U odds to infer total goals, then 1X2 for split.
func lambdaFromW3WithTotals(w3Home, w3Draw, w3Away, overOdds, underOdds, line float64) (LambdaPair, float64) {
	if overOdds <= 0 || underOdds <= 0 || line <= 0 {
		return lambdaFromW3(w3Home, w3Draw, w3Away)
	}
	pOver, _ := removeVigPair(overOdds, underOdds)
	impliedTotal := impliedTotalFromTotals(pOver, line)
	impliedTotal = math.Max(w3ImpliedTotalMin, math.Min(w3ImpliedTotalMax, impliedTotal))
	return splitLambdaBy1X2(w3Home, w3Away, impliedTotal), impliedTotal
}

func splitLambdaBy1X2(w3Home, w3Away, impliedTotal float64) LambdaPair {
	denom := w3Home + w3Away
	if denom < 1e-9 {
		half := impliedTotal / 2
		return LambdaPair{Home: half, Away: half}
	}
	shareHome := w3Home / denom
	return LambdaPair{
		Home: impliedTotal * shareHome,
		Away: impliedTotal * (1 - shareHome),
	}
}

// impliedTotalFromTotals inverts Poisson P(over line) to a total-goals rate λ.
func impliedTotalFromTotals(pOver, line float64) float64 {
	pOver = math.Max(0.05, math.Min(0.95, pOver))
	lo, hi := 0.5, 5.5
	for i := 0; i < 64; i++ {
		mid := (lo + hi) / 2
		if poissonOverProb(mid, line) > pOver {
			hi = mid
		} else {
			lo = mid
		}
	}
	return (lo + hi) / 2
}

func poissonOverProb(lambda, line float64) float64 {
	// Half-integer lines only (e.g. 2.5 → need 3+ goals).
	threshold := int(math.Floor(line)) + 1
	probUnder := 0.0
	for k := 0; k < threshold; k++ {
		probUnder += poissonPMF(lambda, k)
	}
	return 1 - probUnder
}

func poissonPMF(lambda float64, k int) float64 {
	if lambda <= 0 {
		if k == 0 {
			return 1
		}
		return 0
	}
	return math.Exp(-lambda) * math.Pow(lambda, float64(k)) / factorial(k)
}

func factorial(n int) float64 {
	if n <= 1 {
		return 1
	}
	out := 1.0
	for i := 2; i <= n; i++ {
		out *= float64(i)
	}
	return out
}

func removeVigPair(aOdds, bOdds float64) (a, b float64) {
	ia := 1 / aOdds
	ib := 1 / bOdds
	total := ia + ib
	if total == 0 {
		return 0.5, 0.5
	}
	return ia / total, ib / total
}

func blendLambda(w teamdata.ModelWeights, l2, l1, l3 LambdaPair) LambdaPair {
	return LambdaPair{
		Home: w.W2*l2.Home + w.W1*l1.Home + w.W3*l3.Home,
		Away: w.W2*l2.Away + w.W1*l1.Away + w.W3*l3.Away,
	}
}

func lambdaContrib(w float64, l LambdaPair) LambdaPair {
	return LambdaPair{Home: w * l.Home, Away: w * l.Away}
}

func lambdaFavor(l LambdaPair) string {
	return outcomeFavor(l.Home, l.Away)
}
