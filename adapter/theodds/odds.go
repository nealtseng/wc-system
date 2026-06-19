// Package theodds fetches FIFA World Cup 2026 h2h odds from The Odds API.
// Docs: https://the-odds-api.com/liveapi/guides/v4/
package theodds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	sportWC2026 = "soccer_fifa_world_cup"
	baseURL     = "https://api.the-odds-api.com/v4"
	httpTimeout = 30 * time.Second
	sourceName  = "the-odds-api"
)

// MatchOdds is the best available h2h line for one fixture.
type MatchOdds struct {
	EventID         string
	CommenceTime    time.Time
	HomeTeam        string
	AwayTeam        string
	Bookmaker       string
	HomeOdds        float64
	DrawOdds        float64
	AwayOdds        float64
	TotalsLine      float64
	OverOdds        float64
	UnderOdds       float64
	TotalsBookmaker string
}

// HasTotals reports whether O/U odds were found for this event.
func (m MatchOdds) HasTotals() bool {
	return m.TotalsLine > 0 && m.OverOdds > 0 && m.UnderOdds > 0
}

// SourceName returns the DB source label for historical_odds.
func SourceName() string { return sourceName }

type rawOutcome struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Point float64 `json:"point"`
}

type rawMarket struct {
	Key      string       `json:"key"`
	Outcomes []rawOutcome `json:"outcomes"`
}

type rawBookmaker struct {
	Key       string      `json:"key"`
	Title     string      `json:"title"`
	Markets   []rawMarket `json:"markets"`
}

type rawEvent struct {
	ID           string         `json:"id"`
	CommenceTime string         `json:"commence_time"`
	HomeTeam     string         `json:"home_team"`
	AwayTeam     string         `json:"away_team"`
	Bookmakers   []rawBookmaker `json:"bookmakers"`
}

// FetchWC2026 retrieves head-to-head decimal odds for upcoming WC2026 matches.
// Returns an error if apiKey is empty or the HTTP call fails.
func FetchWC2026(ctx context.Context, apiKey string) ([]MatchOdds, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("theodds: API key is empty")
	}

	url := fmt.Sprintf("%s/sports/%s/odds", baseURL, sportWC2026)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("theodds: new request: %w", err)
	}
	q := req.URL.Query()
	q.Set("apiKey", apiKey)
	q.Set("regions", "eu")
	q.Set("markets", "h2h,totals")
	q.Set("oddsFormat", "decimal")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req) //nolint:gosec // fixed API host
	if err != nil {
		return nil, fmt.Errorf("theodds: GET odds: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("theodds: status %d: %s", resp.StatusCode, body)
	}

	var events []rawEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("theodds: decode: %w", err)
	}

	out := make([]MatchOdds, 0, len(events))
	for _, ev := range events {
		commence, err := time.Parse(time.RFC3339, ev.CommenceTime)
		if err != nil {
			continue
		}
		home, draw, away, bookmaker, ok := extractH2H(ev.HomeTeam, ev.AwayTeam, ev.Bookmakers)
		if !ok {
			continue
		}
		tLine, tOver, tUnder, tBook := extractTotals(ev.Bookmakers)
		out = append(out, MatchOdds{
			EventID:         ev.ID,
			CommenceTime:    commence.UTC(),
			HomeTeam:        ev.HomeTeam,
			AwayTeam:        ev.AwayTeam,
			Bookmaker:       bookmaker,
			HomeOdds:        home,
			DrawOdds:        draw,
			AwayOdds:        away,
			TotalsLine:      tLine,
			OverOdds:        tOver,
			UnderOdds:       tUnder,
			TotalsBookmaker: tBook,
		})
	}
	return out, nil
}

func extractH2H(homeTeam, awayTeam string, bookmakers []rawBookmaker) (home, draw, away float64, bookmaker string, ok bool) {
	for _, bm := range bookmakers {
		for _, mkt := range bm.Markets {
			if mkt.Key != "h2h" {
				continue
			}
			var h, d, a float64
			var foundH, foundD, foundA bool
			for _, o := range mkt.Outcomes {
				switch strings.EqualFold(o.Name, homeTeam) {
				case true:
					h, foundH = o.Price, true
				default:
					switch strings.EqualFold(o.Name, awayTeam) {
					case true:
						a, foundA = o.Price, true
					default:
						if strings.EqualFold(o.Name, "Draw") {
							d, foundD = o.Price, true
						}
					}
				}
			}
			if foundH && foundD && foundA && h > 0 && d > 0 && a > 0 {
				return h, d, a, bm.Title, true
			}
		}
	}
	return 0, 0, 0, "", false
}

type totalsCandidate struct {
	line  float64
	over  float64
	under float64
	book  string
	pri   int
}

var totalsBookPriority = map[string]int{
	"pinnacle":     0,
	"betonline.ag": 1,
	"matchbook":    2,
	"1xbet":        3,
	"williamhill":  4,
}

func extractTotals(bookmakers []rawBookmaker) (line, over, under float64, bookmaker string) {
	var at25 []totalsCandidate
	var anyLine []totalsCandidate
	for _, bm := range bookmakers {
		pri := 50
		if p, ok := totalsBookPriority[strings.ToLower(bm.Key)]; ok {
			pri = p
		}
		for _, mkt := range bm.Markets {
			if mkt.Key != "totals" {
				continue
			}
			var pt, ov, un float64
			var hasOver, hasUnder bool
			for _, o := range mkt.Outcomes {
				switch strings.EqualFold(o.Name, "Over") {
				case true:
					if o.Point > 0 && o.Price > 0 {
						pt, ov, hasOver = o.Point, o.Price, true
					}
				default:
					if strings.EqualFold(o.Name, "Under") && o.Price > 0 {
						un, hasUnder = o.Price, true
					}
				}
			}
			if !hasOver || !hasUnder {
				continue
			}
			c := totalsCandidate{line: pt, over: ov, under: un, book: bm.Title, pri: pri}
			anyLine = append(anyLine, c)
			if pt == 2.5 {
				at25 = append(at25, c)
			}
		}
	}
	pick := func(list []totalsCandidate) (totalsCandidate, bool) {
		if len(list) == 0 {
			return totalsCandidate{}, false
		}
		best := list[0]
		for _, c := range list[1:] {
			if c.pri < best.pri {
				best = c
			}
		}
		return best, true
	}
	if c, ok := pick(at25); ok {
		return c.line, c.over, c.under, c.book
	}
	if c, ok := pick(anyLine); ok {
		return c.line, c.over, c.under, c.book
	}
	return 0, 0, 0, ""
}

// NormalizeName lowercases and applies common alias mapping for team matching.
func NormalizeName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	aliases := map[string]string{
		"united states":                 "usa",
		"usa":                           "usa",
		"korea republic":                "south korea",
		"south korea":                   "south korea",
		"turkey":                        "turkiye",
		"turkiye":                       "turkiye",
		"ivory coast":                   "ivory coast",
		"côte d'ivoire":                 "ivory coast",
		"cote d'ivoire":                 "ivory coast",
		"democratic republic of the congo": "congo dr",
		"dr congo":                      "congo dr",
		"congo dr":                      "congo dr",
		"curacao":                       "curaçao",
		"curaçao":                       "curaçao",
		"bosnia and herzegovina":        "bosnia and herzegovina",
		"bosnia & herzegovina":          "bosnia and herzegovina",
	}
	if canon, ok := aliases[n]; ok {
		return canon
	}
	return n
}

// TeamsMatch reports whether two team labels refer to the same side after normalization.
func TeamsMatch(a, b string) bool {
	return NormalizeName(a) == NormalizeName(b)
}
