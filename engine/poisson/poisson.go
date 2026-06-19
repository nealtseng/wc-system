// Package poisson implements a Monte Carlo Poisson simulation engine for
// forecasting football match outcomes.
//
// Each call to Simulate runs `Iterations` independent trials, drawing from
// two independent Poisson distributions parameterised by LambdaHome and
// LambdaAway.  The result is the empirical frequency of each outcome.
package poisson

import (
	"math"
	"math/rand"
	"sort"
)

// homeAdvantageFactor multiplies the home team's base lambda to model the
// well-documented empirical home field advantage in football.
const homeAdvantageFactor = 1.1

// eloScalingExponent controls how strongly the ELO ratio adjusts expected goals.
const eloScalingExponent = 0.3

// DefaultBaseXG is the fallback expected goals per match when team stats are unavailable.
const DefaultBaseXG = 1.2

// PoissonInput parameterises a single simulation run.
type PoissonInput struct {
	LambdaHome float64
	LambdaAway float64
	Iterations int
	// Seed for the PRNG. 0 uses the default reproducible seed (42).
	Seed int64
}

// ScoreProbability represents the empirical probability of a specific scoreline.
type ScoreProbability struct {
	Home int
	Away int
	Prob float64
}

// PoissonResult contains the aggregated output of a Simulate call.
type PoissonResult struct {
	HomeWin    float64
	Draw       float64
	AwayWin    float64
	LambdaHome float64
	LambdaAway float64
	TopScores  []ScoreProbability
}

// ComputeLambdas calculates the adjusted expected-goals lambdas from raw xG
// values and the two teams' ELO ratings.
//
//	lambda_home = baseXGHome * (eloHome/eloAway)^0.3 [* homeAdvantageFactor when not neutral]
//	lambda_away = baseXGAway * (eloAway/eloHome)^0.3
func ComputeLambdas(baseXGHome, baseXGAway, eloHome, eloAway float64, isNeutral bool) (lambdaHome, lambdaAway float64) {
	ratio := eloHome / eloAway
	lambdaHome = baseXGHome * math.Pow(ratio, eloScalingExponent)
	lambdaAway = baseXGAway * math.Pow(1.0/ratio, eloScalingExponent)
	if !isNeutral {
		lambdaHome *= homeAdvantageFactor
	}
	return
}

// poissonSample draws a random integer from a Poisson distribution with the
// given lambda using Knuth's algorithm.  This avoids importing gonum for the
// Monte Carlo step so the package compiles without CGO.
func poissonSample(lambda float64, rng *rand.Rand) int {
	L := math.Exp(-lambda)
	k := 0
	p := 1.0
	for p > L {
		k++
		p *= rng.Float64()
	}
	return k - 1
}

// scoreKey encodes a (home, away) pair into an int64 for use as a map key.
// Supports scores up to 99 goals per side, which is more than sufficient.
func scoreKey(home, away int) int64 {
	return int64(home)*100 + int64(away)
}

// Simulate runs the Monte Carlo Poisson simulation and returns aggregated
// outcome probabilities together with the top-4 most likely scorelines.
func Simulate(input PoissonInput) PoissonResult {
	if input.Iterations <= 0 {
		input.Iterations = 10000
	}

	seed := input.Seed
	if seed == 0 {
		seed = 42
	}
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // seed is caller-controlled

	var homeWins, draws, awayWins int
	scoreCounts := make(map[int64]int)

	for i := 0; i < input.Iterations; i++ {
		h := poissonSample(input.LambdaHome, rng)
		a := poissonSample(input.LambdaAway, rng)

		scoreCounts[scoreKey(h, a)]++

		switch {
		case h > a:
			homeWins++
		case h == a:
			draws++
		default:
			awayWins++
		}
	}

	total := float64(input.Iterations)

	// Build top-4 scorelines sorted by descending probability.
	type kv struct {
		key   int64
		count int
	}
	pairs := make([]kv, 0, len(scoreCounts))
	for k, v := range scoreCounts {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	maxTop := 4
	if len(pairs) < maxTop {
		maxTop = len(pairs)
	}
	topScores := make([]ScoreProbability, maxTop)
	for i := 0; i < maxTop; i++ {
		h := int(pairs[i].key / 100)
		a := int(pairs[i].key % 100)
		topScores[i] = ScoreProbability{
			Home: h,
			Away: a,
			Prob: float64(pairs[i].count) / total,
		}
	}

	return PoissonResult{
		HomeWin:    float64(homeWins) / total,
		Draw:       float64(draws) / total,
		AwayWin:    float64(awayWins) / total,
		LambdaHome: input.LambdaHome,
		LambdaAway: input.LambdaAway,
		TopScores:  topScores,
	}
}
