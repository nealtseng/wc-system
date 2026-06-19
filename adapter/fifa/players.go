// Package fifa loads EA Sports FC player rosters from a public CSV mirror
// (SoFIFA-derived dataset on GitHub; same family as Kaggle "FIFA Player Data").
package fifa

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultCSVURL = "https://raw.githubusercontent.com/SolideSpoke/sofifa-web-scraper/main/output/player-data-full.csv"

// Player is one national-team squad member normalised for the API.
type Player struct {
	ID        string
	No        int
	Name      string
	WikiSlug  string
	Pos       string
	Age       int
	Fitness   int
	Value     string
	Prof      int
	Imp       int
	Role      string
	OffPitch  float64
	ImageURL  string
	Rating    int // country_rating for sorting
}

// FetchSquads downloads the FIFA CSV and builds squads keyed by FIFA country_name.
func FetchSquads(ctx context.Context, csvURL string, countryNames map[string]string) (map[string][]Player, error) {
	if csvURL == "" {
		csvURL = defaultCSVURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, csvURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fifa: build request: %w", err)
	}

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req) //nolint:gosec // fixed public dataset URL
	if err != nil {
		return nil, fmt.Errorf("fifa: download csv: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fifa: unexpected status %d", resp.StatusCode)
	}

	return ParseSquads(resp.Body, countryNames)
}

// ParseSquads reads CSV rows and groups national-team players by country name.
func ParseSquads(r io.Reader, countryNames map[string]string) (map[string][]Player, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("fifa: read header: %w", err)
	}
	col := indexColumns(header)

	required := []string{"name", "dob", "overall_rating", "potential", "value", "country_name", "country_kit_number", "country_position", "country_rating", "stamina", "international_reputation", "specialities", "image"}
	for _, key := range required {
		if col[key] < 0 {
			return nil, fmt.Errorf("fifa: missing column %q", key)
		}
	}

	allowed := make(map[string]struct{}, len(countryNames))
	for _, name := range countryNames {
		allowed[name] = struct{}{}
	}

	byCountry := make(map[string]map[string]Player)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("fifa: read row: %w", err)
		}
		if len(row) <= col["country_name"] {
			continue
		}

		country := strings.TrimSpace(row[col["country_name"]])
		if country == "" {
			continue
		}
		if _, ok := allowed[country]; !ok {
			continue
		}

		playerID := strings.TrimSpace(safeCol(row, col, "player_id"))
		if playerID == "" {
			continue
		}

		name := strings.TrimSpace(row[col["name"]])
		countryRating, _ := strconv.Atoi(strings.TrimSpace(row[col["country_rating"]]))
		overall, _ := strconv.Atoi(strings.TrimSpace(row[col["overall_rating"]]))
		potential, _ := strconv.Atoi(strings.TrimSpace(row[col["potential"]]))
		stamina, _ := strconv.Atoi(strings.TrimSpace(row[col["stamina"]]))
		rep, _ := strconv.Atoi(strings.TrimSpace(row[col["international_reputation"]]))
		kitNo, _ := strconv.Atoi(strings.TrimSpace(row[col["country_kit_number"]]))

		pos := strings.TrimSpace(row[col["country_position"]])
		if pos == "" {
			pos = firstPosition(strings.TrimSpace(safeCol(row, col, "positions")))
		}

		role := pos
		if spec := strings.TrimSpace(row[col["specialities"]]); spec != "" {
			parts := strings.Split(spec, ",")
			if len(parts) > 0 {
				role = pos + " — " + strings.TrimSpace(parts[0])
			}
		}

		p := Player{
			ID:       playerID,
			No:       kitNo,
			Name:     name,
			WikiSlug: wikiSlug(name),
			Pos:      pos,
			Age:      ageFromDOB(strings.TrimSpace(row[col["dob"]])),
			Fitness:  clamp(stamina, 0, 100),
			Value:    strings.TrimSpace(row[col["value"]]),
			Prof:     clamp(int(math.Round(float64(overall)/5.0)), 0, 20),
			Imp:      clamp(int(math.Round(float64(potential)/5.0)), 0, 20),
			Role:     role,
			OffPitch: float64(clamp(rep*2, 2, 10)),
			ImageURL: strings.TrimSpace(safeCol(row, col, "image")),
			Rating:   countryRating,
		}

		if byCountry[country] == nil {
			byCountry[country] = make(map[string]Player)
		}
		existing, ok := byCountry[country][playerID]
		if !ok || p.Rating > existing.Rating {
			byCountry[country][playerID] = p
		}
	}

	out := make(map[string][]Player, len(countryNames))
	for teamID, country := range countryNames {
		playersMap := byCountry[country]
		players := make([]Player, 0, len(playersMap))
		for _, p := range playersMap {
			players = append(players, p)
		}
		sort.Slice(players, func(i, j int) bool {
			if players[i].Rating != players[j].Rating {
				return players[i].Rating > players[j].Rating
			}
			return players[i].Prof > players[j].Prof
		})
		if len(players) > 26 {
			players = players[:26]
		}
		out[teamID] = players
	}

	return out, nil
}

func indexColumns(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		key := strings.Trim(strings.TrimSpace(h), `"`)
		idx[key] = i
	}
	return idx
}

func safeCol(row []string, col map[string]int, name string) string {
	i := col[name]
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}

func firstPosition(positions string) string {
	if positions == "" {
		return "—"
	}
	parts := strings.Split(positions, ",")
	return strings.TrimSpace(parts[0])
}

func wikiSlug(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}

func ageFromDOB(dob string) int {
	if dob == "" {
		return 0
	}
	t, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - t.Year()
	if now.YearDay() < t.YearDay() {
		age--
	}
	return age
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
