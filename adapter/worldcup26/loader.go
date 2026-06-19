// Package worldcup26 fetches 2026 World Cup fixtures from the public GitHub raw CDN.
package worldcup26

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	rawTeamsURL    = "https://raw.githubusercontent.com/rezarahiminia/worldcup2026/main/football.teams.json"
	rawStadiumsURL = "https://raw.githubusercontent.com/rezarahiminia/worldcup2026/main/football.stadiums.json"
	rawMatchesURL  = "https://raw.githubusercontent.com/rezarahiminia/worldcup2026/main/football.matches.json"
	httpTimeout    = 30 * time.Second
)

// RawTeam is one national team entry from football.teams.json.
type RawTeam struct {
	ID       string `json:"id"`
	NameEN   string `json:"name_en"`
	FIFACode string `json:"fifa_code"`
	Groups   string `json:"groups"`
	Flag     string `json:"flag"`
}

// RawStadium is one venue entry from football.stadiums.json.
type RawStadium struct {
	ID         string `json:"id"`
	NameEN     string `json:"name_en"`
	FIFAName   string `json:"fifa_name"`
	CityEN     string `json:"city_en"`
	CountryEN  string `json:"country_en"`
	Capacity   int    `json:"capacity"`
}

// RawMatch is one fixture entry from football.matches.json.
type RawMatch struct {
	ID           string `json:"id"`
	HomeTeamID   string `json:"home_team_id"`
	AwayTeamID   string `json:"away_team_id"`
	HomeScore    string `json:"home_score"`
	AwayScore    string `json:"away_score"`
	Group        string `json:"group"`
	Matchday     string `json:"matchday"`
	LocalDate    string `json:"local_date"`
	StadiumID    string `json:"stadium_id"`
	Finished     string `json:"finished"`
	TimeElapsed  string `json:"time_elapsed"`
	Type         string `json:"type"`
	HomeNameEN   string `json:"home_team_name_en"`
	AwayNameEN   string `json:"away_team_name_en"`
	HomeLabel    string `json:"home_team_label"`
	AwayLabel    string `json:"away_team_label"`
}

// FetchTeams downloads and parses the teams JSON array.
func FetchTeams(ctx context.Context) ([]RawTeam, error) {
	return fetch[RawTeam](ctx, rawTeamsURL)
}

// FetchStadiums downloads and parses the stadiums JSON array.
func FetchStadiums(ctx context.Context) ([]RawStadium, error) {
	return fetch[RawStadium](ctx, rawStadiumsURL)
}

// FetchMatches downloads and parses the matches JSON array.
func FetchMatches(ctx context.Context) ([]RawMatch, error) {
	return fetch[RawMatch](ctx, rawMatchesURL)
}

// BuildStadiumMap maps numeric stadium IDs to English venue names.
func BuildStadiumMap(stadiums []RawStadium) map[string]string {
	out := make(map[string]string, len(stadiums))
	for _, s := range stadiums {
		out[s.ID] = s.NameEN
	}
	return out
}

// BuildWCIDToFIFAMap maps numeric team IDs to FIFA catalog codes.
func BuildWCIDToFIFAMap(teams []RawTeam) map[string]string {
	out := make(map[string]string, len(teams))
	for _, t := range teams {
		out[t.ID] = t.FIFACode
	}
	return out
}

func fetch[T any](ctx context.Context, url string) ([]T, error) {
	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("worldcup26: new request: %w", err)
	}

	resp, err := client.Do(req) //nolint:gosec // URL is a fixed public CDN endpoint
	if err != nil {
		return nil, fmt.Errorf("worldcup26: GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("worldcup26: GET %s: status %d: %s", url, resp.StatusCode, body)
	}

	var result []T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("worldcup26: decode %s: %w", url, err)
	}
	return result, nil
}
