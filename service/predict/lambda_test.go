package predict

import (
	"math"
	"testing"

	"wc-system/service/teamdata"
)

func TestLambdaFromW3DrawLowersTotal(t *testing.T) {
	_, totalHigh := lambdaFromW3(0.25, 0.35, 0.40)
	_, totalLow := lambdaFromW3(0.40, 0.20, 0.40)
	if totalHigh >= totalLow {
		t.Fatalf("higher draw should imply lower total: %v vs %v", totalHigh, totalLow)
	}
}

func TestLambdaFromW1SkewsToStrongerSide(t *testing.T) {
	base := LambdaPair{Home: 1.3, Away: 1.3}
	strongHome := lambdaFromW1(base, 0.85, 0.55)
	if strongHome.Home <= strongHome.Away {
		t.Fatalf("stronger home should get higher lambda: %v vs %v", strongHome.Home, strongHome.Away)
	}
}

func TestBlendLambdaWeights(t *testing.T) {
	w := teamdata.ModelWeights{W1: 0.3, W2: 0.3, W3: 0.4}
	l2 := LambdaPair{Home: 1.0, Away: 1.0}
	l1 := LambdaPair{Home: 1.2, Away: 0.8}
	l3 := LambdaPair{Home: 0.8, Away: 1.2}
	got := blendLambda(w, l2, l1, l3)
	wantH := 0.3*1.0 + 0.3*1.2 + 0.4*0.8
	wantA := 0.3*1.0 + 0.3*0.8 + 0.4*1.2
	if math.Abs(got.Home-wantH) > 1e-9 || math.Abs(got.Away-wantA) > 1e-9 {
		t.Fatalf("blend = %v, want %v/%v", got, wantH, wantA)
	}
}

func TestImpliedTotalFromTotalsNearLine(t *testing.T) {
	// Equal O/U at 2.5 → implied total should be close to 2.5
	got := impliedTotalFromTotals(0.5, 2.5)
	if math.Abs(got-2.65) > 0.1 {
		t.Fatalf("equal O/U at 2.5: implied total = %.3f, want ~2.65 (Poisson inversion)", got)
	}
}

func TestLambdaFromW3WithTotalsUsesOverBias(t *testing.T) {
	proxy, _ := lambdaFromW3(0.45, 0.25, 0.30)
	withTotals, total := lambdaFromW3WithTotals(0.45, 0.25, 0.30, 1.65, 2.35, 2.5)
	if total <= proxy.Home+proxy.Away {
		t.Fatalf("over-favored line should raise total: totals=%.2f proxy=%.2f", total, proxy.Home+proxy.Away)
	}
	if withTotals.Home <= withTotals.Away {
		t.Fatalf("home-favored 1X2 should skew lambdas home: %v", withTotals)
	}
}
