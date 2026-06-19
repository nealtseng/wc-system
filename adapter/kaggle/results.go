// Package kaggle loads international match results from the martj42 dataset
// (mirrored on GitHub; same data as kaggle.com/datasets/martj42/international-football-results-from-1872-to-2024).
package kaggle

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const resultsCSVURL = "https://raw.githubusercontent.com/martj42/international_results/master/results.csv"

// Match is one row from the international results dataset.
type Match struct {
	Date      time.Time
	HomeTeam  string
	AwayTeam  string
	HomeScore int
	AwayScore int
	Tournament string
	Neutral   bool
}

// TeamStats aggregates derived metrics for one national team.
type TeamStats struct {
	ELO             float64
	AvgGoalsFor     float64
	AvgGoalsAgainst float64
	WinRate         float64
	Momentum        float64 // recent form score in [-1, 1]
	MatchesPlayed   int
}

// FetchResults downloads and parses the martj42 CSV, keeping only matches
// that involve at least one team in allowedNames.
func FetchResults(allowedNames map[string]struct{}) ([]Match, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(resultsCSVURL) //nolint:gosec // fixed public dataset URL
	if err != nil {
		return nil, fmt.Errorf("kaggle: download results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kaggle: unexpected status %d", resp.StatusCode)
	}

	return ParseResults(resp.Body, allowedNames)
}

// ParseResults reads the CSV body and filters to allowed team names.
func ParseResults(r io.Reader, allowedNames map[string]struct{}) ([]Match, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("kaggle: read header: %w", err)
	}
	col := indexColumns(header)

	var matches []Match
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("kaggle: read row: %w", err)
		}
		if len(row) <= col.awayScore {
			continue
		}

		home := strings.TrimSpace(row[col.homeTeam])
		away := strings.TrimSpace(row[col.awayTeam])
		if _, ok := allowedNames[home]; !ok {
			if _, ok := allowedNames[away]; !ok {
				continue
			}
		}

		date, err := time.Parse("2006-01-02", strings.TrimSpace(row[col.date]))
		if err != nil {
			continue
		}

		homeScore, _ := strconv.Atoi(strings.TrimSpace(row[col.homeScore]))
		awayScore, _ := strconv.Atoi(strings.TrimSpace(row[col.awayScore]))

		neutral := false
		if col.neutral >= 0 && col.neutral < len(row) {
			neutral = strings.EqualFold(strings.TrimSpace(row[col.neutral]), "True")
		}

		tournament := ""
		if col.tournament >= 0 && col.tournament < len(row) {
			tournament = strings.TrimSpace(row[col.tournament])
		}

		matches = append(matches, Match{
			Date:       date,
			HomeTeam:   home,
			AwayTeam:   away,
			HomeScore:  homeScore,
			AwayScore:  awayScore,
			Tournament: tournament,
			Neutral:    neutral,
		})
	}

	return matches, nil
}

type columnIndex struct {
	date, homeTeam, awayTeam, homeScore, awayScore, tournament, neutral int
}

func indexColumns(header []string) columnIndex {
	idx := columnIndex{
		date: -1, homeTeam: -1, awayTeam: -1, homeScore: -1, awayScore: -1, tournament: -1, neutral: -1,
	}
	for i, h := range header {
		switch strings.TrimSpace(h) {
		case "date":
			idx.date = i
		case "home_team":
			idx.homeTeam = i
		case "away_team":
			idx.awayTeam = i
		case "home_score":
			idx.homeScore = i
		case "away_score":
			idx.awayScore = i
		case "tournament":
			idx.tournament = i
		case "neutral":
			idx.neutral = i
		}
	}
	return idx
}

type teamAgg struct {
	goalsFor, goalsAgainst, wins, draws, played int
	recentPoints                                []float64
}

// ComputeStats derives ELO ratings and goal averages from chronological matches.
func ComputeStats(matches []Match, teamNames []string) map[string]TeamStats {
	tracked := make(map[string]struct{}, len(teamNames))
	ratings := make(map[string]float64, len(teamNames))
	for _, name := range teamNames {
		tracked[name] = struct{}{}
		ratings[name] = 1500
	}

	byTeam := make(map[string]*teamAgg, len(teamNames))
	for _, name := range teamNames {
		byTeam[name] = &teamAgg{}
	}

	const homeAdv = 100.0
	const kFactor = 40.0
	cutoff := time.Now().AddDate(-3, 0, 0)

	for _, m := range matches {
		homeScore := scoreForHome(m.HomeScore, m.AwayScore)
		rHome, homeTracked := ratings[m.HomeTeam]
		if !homeTracked {
			rHome = 1500
		}
		rAway, awayTracked := ratings[m.AwayTeam]
		if !awayTracked {
			rAway = 1500
		}

		effectiveHome := rHome
		if !m.Neutral {
			effectiveHome += homeAdv
		}
		expected := 1.0 / (1.0 + math.Pow(10, (rAway-effectiveHome)/400.0))
		delta := kFactor * (homeScore - expected)

		if homeTracked {
			ratings[m.HomeTeam] = rHome + delta
		}
		if awayTracked {
			ratings[m.AwayTeam] = rAway - delta
		}

		if homeTracked {
			updateAgg(byTeam, m.HomeTeam, m.HomeScore, m.AwayScore, homeScore, m.Date.After(cutoff))
		}
		if awayTracked {
			updateAgg(byTeam, m.AwayTeam, m.AwayScore, m.HomeScore, 1.0-homeScore, m.Date.After(cutoff))
		}
	}

	out := make(map[string]TeamStats, len(teamNames))
	for _, name := range teamNames {
		a := byTeam[name]
		if a == nil || a.played == 0 {
			elo := ratings[name]
			if elo == 0 {
				elo = 1500
			}
			out[name] = TeamStats{ELO: elo, AvgGoalsFor: 1.2, AvgGoalsAgainst: 1.2}
			continue
		}

		winRate := float64(a.wins) / float64(a.played)
		momentum := 0.0
		if len(a.recentPoints) > 0 {
			sum := 0.0
			for _, p := range a.recentPoints {
				sum += p
			}
			momentum = sum/float64(len(a.recentPoints)) - 0.5 // centre around 0
		}

		out[name] = TeamStats{
			ELO:             ratings[name],
			AvgGoalsFor:     float64(a.goalsFor) / float64(a.played),
			AvgGoalsAgainst: float64(a.goalsAgainst) / float64(a.played),
			WinRate:         winRate,
			Momentum:        momentum,
			MatchesPlayed:   a.played,
		}
	}
	return out
}

func scoreForHome(home, away int) float64 {
	switch {
	case home > away:
		return 1.0
	case home < away:
		return 0.0
	default:
		return 0.5
	}
}

func updateAgg(byTeam map[string]*teamAgg, team string, gf, ga int, matchPoints float64, recent bool) {
	a, ok := byTeam[team]
	if !ok {
		return
	}
	a.goalsFor += gf
	a.goalsAgainst += ga
	a.played++
	switch matchPoints {
	case 1.0:
		a.wins++
	case 0.5:
		a.draws++
	}
	if recent {
		a.recentPoints = append(a.recentPoints, matchPoints)
		if len(a.recentPoints) > 10 {
			a.recentPoints = a.recentPoints[len(a.recentPoints)-10:]
		}
	}
}
