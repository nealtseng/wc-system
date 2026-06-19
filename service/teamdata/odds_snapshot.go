package teamdata

// OddsSnapshot holds the most recent bookmaker odds for one fixture.
type OddsSnapshot struct {
	HomeOdds   float64
	DrawOdds   float64
	AwayOdds   float64
	TotalsLine float64
	OverOdds   float64
	UnderOdds  float64
}

// HasTotals reports whether O/U odds are present.
func (s OddsSnapshot) HasTotals() bool {
	return s.TotalsLine > 0 && s.OverOdds > 0 && s.UnderOdds > 0
}
