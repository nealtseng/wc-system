// Package fbref scrapes national-team xG from FBref match logs.
package fbref

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	fbrefBaseURL     = "https://fbref.com/en/national/"
	fbrefHomeURL     = "https://fbref.com/"
	fbrefDelayMin    = 8 * time.Second
	fbrefDelayMax    = 15 * time.Second
	fbrefMaxRetries  = 3
	fbrefBackoffBase = 8 * time.Second
)

var (
	// ErrNoData is returned when no relevant international xG rows can be parsed.
	ErrNoData = errors.New("fbref: no international xG data")
)

// XGResult is the aggregated xG per match for one national team.
type XGResult struct {
	Avg        float64
	MatchCount int
}

var fbrefUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
}

// DelayBetweenTeams returns a random pause between team page fetches.
func DelayBetweenTeams() time.Duration {
	if fbrefDelayMax <= fbrefDelayMin {
		return fbrefDelayMin
	}
	return fbrefDelayMin + time.Duration(rand.Int63n(int64(fbrefDelayMax-fbrefDelayMin)))
}

// Client reuses one user agent across a batch of team fetches.
type Client struct {
	ua     string
	warmed bool
}

// NewClient builds a FBref scraper session with a fixed desktop user agent.
func NewClient() *Client {
	return &Client{ua: fbrefUserAgents[rand.Intn(len(fbrefUserAgents))]}
}

func (cl *Client) newCollector() *colly.Collector {
	c := colly.NewCollector(colly.UserAgent(cl.ua))
	c.SetRequestTimeout(45 * time.Second)
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Referer", fbrefHomeURL)
	})
	return c
}

// Warmup loads the FBref homepage before team pages (same UA for the batch).
func (cl *Client) Warmup(ctx context.Context) error {
	if cl.warmed {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := cl.newCollector().Visit(fbrefHomeURL); err != nil && !isRetryableStatus(err) {
		return fmt.Errorf("fbref: warmup: %w", err)
	}
	cl.warmed = true
	return nil
}

// FetchNationalTeamXG returns average xG per international match for teamSlug.
// teamSlug is the FBref path suffix, e.g. "ENG/England-Men-Stats".
func FetchNationalTeamXG(ctx context.Context, teamSlug string) (XGResult, error) {
	cl := NewClient()
	if err := cl.Warmup(ctx); err != nil {
		return XGResult{}, err
	}
	return cl.FetchTeamXG(ctx, teamSlug)
}

// FetchTeamXG scrapes one team page using the shared client session.
func (cl *Client) FetchTeamXG(ctx context.Context, teamSlug string) (XGResult, error) {
	url := fbrefBaseURL + teamSlug
	var lastErr error
	for attempt := 0; attempt < fbrefMaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return XGResult{}, err
		}
		if attempt > 0 {
			time.Sleep(fbrefBackoffBase * time.Duration(1<<uint(attempt-1)))
		}
		result, err := cl.fetchOnce(ctx, url)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isRetryableStatus(err) {
			break
		}
	}
	return XGResult{}, lastErr
}

func (cl *Client) fetchOnce(ctx context.Context, url string) (XGResult, error) {
	if err := ctx.Err(); err != nil {
		return XGResult{}, err
	}

	var (
		mu       sync.Mutex
		xgVals   []float64
		fetchErr error
	)

	c := cl.newCollector()

	attempt := 0
	c.OnError(func(r *colly.Response, err error) {
		if r != nil && isRetryableHTTP(r.StatusCode) && attempt < fbrefMaxRetries-1 {
			attempt++
			sleep := fbrefBackoffBase * time.Duration(1<<uint(attempt-1))
			time.Sleep(sleep)
			_ = r.Request.Retry()
			return
		}
		mu.Lock()
		if fetchErr == nil {
			if r != nil {
				fetchErr = fmt.Errorf("fbref: HTTP %d %s: %w", r.StatusCode, r.Request.URL, err)
			} else {
				fetchErr = fmt.Errorf("fbref: request %s: %w", url, err)
			}
		}
		mu.Unlock()
	})

	parseRow := func(e *colly.HTMLElement) {
		comp := strings.ToLower(strings.TrimSpace(e.ChildText(`td[data-stat="comp"]`)))
		if comp == "" {
			comp = strings.ToLower(strings.TrimSpace(e.ChildText(`td[data-stat="competition"]`)))
		}
		if !isInternationalMatchComp(comp) {
			return
		}

		xgText := strings.TrimSpace(e.ChildText(`td[data-stat="xg"]`))
		if xgText == "" || xgText == "Matches" {
			return
		}
		xg, err := strconv.ParseFloat(xgText, 64)
		if err != nil || math.IsNaN(xg) || xg < 0 {
			return
		}
		mu.Lock()
		xgVals = append(xgVals, xg)
		mu.Unlock()
	}

	c.OnHTML(`table#matchlogs_for tbody tr`, parseRow)
	c.OnHTML(`table.stats_table tbody tr`, func(e *colly.HTMLElement) {
		if e.Attr("class") == "thead" || strings.Contains(e.Attr("class"), "spacer") {
			return
		}
		parseRow(e)
	})

	if err := c.Visit(url); err != nil {
		if isRetryableStatus(err) {
			return XGResult{}, err
		}
		return XGResult{}, fmt.Errorf("fbref: visit %s: %w", url, err)
	}

	mu.Lock()
	defer mu.Unlock()
	if fetchErr != nil {
		return XGResult{}, fetchErr
	}
	if len(xgVals) == 0 {
		return XGResult{}, ErrNoData
	}

	sum := 0.0
	for _, v := range xgVals {
		sum += v
	}
	return XGResult{
		Avg:        sum / float64(len(xgVals)),
		MatchCount: len(xgVals),
	}, nil
}

func isInternationalMatchComp(comp string) bool {
	comp = strings.ToLower(strings.TrimSpace(comp))
	if comp == "" {
		return false
	}
	keywords := []string{
		"world cup", "nations league", "friendly", "friendlies",
		"uefa euro", "euro qualif", "copa america", "afcon", "africa cup",
		"asian cup", "gold cup", "qualif", "international",
		"concacaf", "conmebol", "afc ", "caf ", "ofc ",
	}
	for _, kw := range keywords {
		if strings.Contains(comp, kw) {
			return true
		}
	}
	return false
}

func isRetryableHTTP(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusForbidden || code == http.StatusServiceUnavailable
}

func isRetryableStatus(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "forbidden")
}
