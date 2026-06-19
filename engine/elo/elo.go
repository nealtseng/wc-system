// Package elo implements the ELO rating system for international football.
// It computes expected win/draw/loss probabilities from two team ELO ratings.
package elo

import "math"

// TournamentType determines the K-factor used when updating ratings.
type TournamentType string

const (
	TournamentWorldCup    TournamentType = "world_cup"
	TournamentContinental TournamentType = "continental"
	TournamentFriendly    TournamentType = "friendly"
	TournamentDefault     TournamentType = "default"
)

// homeAdvantage is the ELO bonus added to the home team in non-neutral venues.
// This is the standard academic value used by World Football ELO Ratings.
const homeAdvantage = 100.0

// KFactor returns the K-factor for a given tournament type.
func KFactor(t TournamentType) float64 {
	switch t {
	case TournamentWorldCup:
		return 60
	case TournamentContinental:
		return 50
	case TournamentFriendly:
		return 20
	default:
		return 40
	}
}

// ELOInput holds the parameters for a single match ELO calculation.
type ELOInput struct {
	TeamA     string
	TeamB     string
	RatingA   float64 // TeamA is treated as the home team when IsNeutral is false
	RatingB   float64
	IsNeutral bool
}

// ELOResult contains the pre-match expected outcome probabilities derived
// from the two teams' ELO ratings.
type ELOResult struct {
	// ExpectedHomeWin is the probability that TeamA (home) wins.
	ExpectedHomeWin float64
	// ExpectedDraw is an estimate of the probability of a draw.
	ExpectedDraw float64
	// ExpectedAwayWin is the probability that TeamB (away) wins.
	ExpectedAwayWin float64
}

// Calculate derives expected outcome probabilities from ELOInput.
//
// Formula:
//
//	E_A = 1 / (1 + 10^((RatingB - effectiveRatingA) / 400))
//
// Draw probability:
//
//	draw = max(0, 1 - |E_A - 0.5| * 2) * 0.3
func Calculate(input ELOInput) ELOResult {
	effectiveA := input.RatingA
	if !input.IsNeutral {
		effectiveA += homeAdvantage
	}

	// Standard ELO expected score for TeamA.
	eA := 1.0 / (1.0 + math.Pow(10, (input.RatingB-effectiveA)/400.0))

	// Estimate draw probability: highest near 50/50, decays toward 0 as one
	// side becomes a strong favourite.
	drawRaw := (1.0 - math.Abs(eA-0.5)*2.0) * 0.3
	if drawRaw < 0 {
		drawRaw = 0
	}

	// Distribute remaining probability proportionally between win/loss.
	remaining := 1.0 - drawRaw
	homeWin := eA * remaining
	awayWin := (1.0 - eA) * remaining

	return ELOResult{
		ExpectedHomeWin: homeWin,
		ExpectedDraw:    drawRaw,
		ExpectedAwayWin: awayWin,
	}
}
