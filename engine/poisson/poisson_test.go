package poisson_test

import (
	"math"
	"testing"

	"wc-system/engine/poisson"
)

func TestComputeLambdasNeutral(t *testing.T) {
	homeN, awayN := poisson.ComputeLambdas(1.5, 1.2, 1800, 1600, true)
	homeH, awayH := poisson.ComputeLambdas(1.5, 1.2, 1800, 1600, false)
	if homeH <= homeN {
		t.Fatalf("home with advantage %v should exceed neutral %v", homeH, homeN)
	}
	if math.Abs(awayH-awayN) > 1e-9 {
		t.Fatalf("away lambda should not change with home advantage: %v vs %v", awayH, awayN)
	}
}

func TestComputeLambdasTypicalInternationalRange(t *testing.T) {
	home, away := poisson.ComputeLambdas(1.4, 1.35, 1700, 1750, true)
	if home < 0.8 || home > 2.2 || away < 0.8 || away > 2.2 {
		t.Fatalf("typical lambdas out of expected intl range: %v vs %v", home, away)
	}
}
