// Package fmscout scrapes FM26 player attributes from FMScout.
package fmscout

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	fmscoutBaseURL    = "https://www.fmscout.com"
	fmscoutDelayMin   = 3 * time.Second
	fmscoutDelayMax   = 8 * time.Second
	fmscoutMaxRetries = 3
	fmscoutBackoffBase = 5 * time.Second
)

var (
	// ErrNoData is returned when no players could be scraped for a nation.
	ErrNoData = errors.New("fmscout: no player data")
)

var fmscoutUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
}

// PlayerAttribute is one FM26 player row with scraped attributes.
type PlayerAttribute struct {
	Name       string
	Position   string
	Overall    float64
	Attributes map[string]int
}

// FetchNationalTeamPlayers returns FM26 players for a national team name, e.g. "France".
func FetchNationalTeamPlayers(ctx context.Context, teamName string) ([]PlayerAttribute, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	searchURL := fmscoutBaseURL + "/a-search-players26.html?n=" + url.QueryEscape(teamName)

	var (
		mu        sync.Mutex
		players   []PlayerAttribute
		profileLinks []string
		fetchErr  error
	)

	c := colly.NewCollector(
		colly.UserAgent(fmscoutUserAgents[rand.Intn(len(fmscoutUserAgents))]),
	)
	c.SetRequestTimeout(30 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
	})

	attempt := 0
	c.OnError(func(r *colly.Response, err error) {
		if r != nil && r.StatusCode == 429 && attempt < fmscoutMaxRetries {
			attempt++
			sleep := fmscoutBackoffBase * time.Duration(1<<uint(attempt-1))
			time.Sleep(sleep)
			_ = r.Request.Retry()
			return
		}
		mu.Lock()
		if fetchErr == nil {
			fetchErr = fmt.Errorf("fmscout: request %s: %w", r.Request.URL, err)
		}
		mu.Unlock()
	})

	c.OnHTML(`table tbody tr`, func(e *colly.HTMLElement) {
		nat := strings.TrimSpace(e.ChildText(`td[data-field="nationality"]`))
		if nat == "" {
			nat = strings.TrimSpace(e.ChildText(`td:nth-child(4)`))
		}
		if nat != "" && !strings.EqualFold(nat, teamName) {
			return
		}

		link := e.DOM.Find(`a[href*="/a-"]`).First()
		href, ok := link.Attr("href")
		if !ok || href == "" {
			return
		}
		if !strings.HasPrefix(href, "http") {
			href = fmscoutBaseURL + href
		}

		mu.Lock()
		profileLinks = append(profileLinks, href)
		mu.Unlock()
	})

	if err := c.Visit(searchURL); err != nil {
		return nil, fmt.Errorf("fmscout: visit search: %w", err)
	}

	mu.Lock()
	links := append([]string(nil), profileLinks...)
	mu.Unlock()

	if fetchErr != nil && len(links) == 0 {
		return nil, fetchErr
	}
	if len(links) == 0 {
		return nil, ErrNoData
	}

	const maxProfiles = 30
	if len(links) > maxProfiles {
		links = links[:maxProfiles]
	}

	pc := c.Clone()
	pc.OnHTML(`table.player_attributes tr, table.attributes tr, .player-attributes tr`, func(e *colly.HTMLElement) {
		// handled per visit below
	})

	for _, profileURL := range links {
		if err := ctx.Err(); err != nil {
			return players, err
		}

		p := PlayerAttribute{Attributes: make(map[string]int)}
		var parseErr error

		visitColly := pc.Clone()
		visitColly.OnHTML("h1", func(e *colly.HTMLElement) {
			p.Name = strings.TrimSpace(e.Text)
		})
		visitColly.OnHTML(`table.player_attributes tr, table.attributes tr`, func(e *colly.HTMLElement) {
			cells := e.DOM.Find("td")
			if cells.Length() < 2 {
				return
			}
			key := strings.TrimSpace(cells.First().Text())
			valText := strings.TrimSpace(cells.Last().Text())
			val, err := strconv.Atoi(valText)
			if err != nil || key == "" {
				return
			}
			p.Attributes[key] = val
		})
		visitColly.OnHTML(`.rating, .overall, [data-field="rating"]`, func(e *colly.HTMLElement) {
			if p.Overall > 0 {
				return
			}
			stars, err := strconv.ParseFloat(strings.TrimSpace(e.Text), 64)
			if err == nil && stars > 0 && stars <= 5 {
				p.Overall = stars * 20
			}
		})

		if err := visitColly.Visit(profileURL); err != nil {
			parseErr = err
		}

		if p.Name == "" {
			continue
		}
		if p.Overall <= 0 && len(p.Attributes) > 0 {
			sum := 0
			for _, v := range p.Attributes {
				sum += v
			}
			p.Overall = float64(sum) / float64(len(p.Attributes)) * 5
		}
		if p.Overall <= 0 || math.IsNaN(p.Overall) {
			if parseErr != nil {
				continue
			}
			continue
		}
		players = append(players, p)
		time.Sleep(fmscoutDelayMin + time.Duration(rand.Int63n(int64(fmscoutDelayMax-fmscoutDelayMin))))
	}

	if len(players) == 0 {
		return nil, ErrNoData
	}
	return players, nil
}
