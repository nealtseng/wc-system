package handlers

// removeVig returns de-vigged implied probabilities from decimal odds.
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

// FractionalKelly computes (p*b - q) / b * scale, floored at 0.
func FractionalKelly(p, decimalOdds, scale float64) float64 {
	if decimalOdds <= 1 {
		return 0
	}
	b := decimalOdds - 1
	q := 1 - p
	kelly := (p*b - q) / b
	if kelly < 0 {
		return 0
	}
	return kelly * scale
}
