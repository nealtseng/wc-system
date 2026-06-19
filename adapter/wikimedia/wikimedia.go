// Package wikimedia fetches national football team metadata from the Wikipedia REST API.
// Endpoint: https://en.wikipedia.org/api/rest_v1/page/summary/{slug}
// No API key required.
package wikimedia

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrNotFound is returned when the Wikipedia page does not exist (HTTP 404).
var ErrNotFound = errors.New("wikimedia: page not found")

// WikiPageSummary contains parsed Wikipedia REST summary fields.
type WikiPageSummary struct {
	Slug         string
	Extract      string
	PageURL      string
	ThumbnailURL string
	FetchedAt    time.Time
}

// WikiTeamMeta contains the parsed metadata for a national football team.
type WikiTeamMeta struct {
	CountrySlug string
	Extract     string // Wikipedia plain-text summary (contains historical records)
	PageURL     string
	FetchedAt   time.Time
}

// FetchPageSummary retrieves summary, page URL, and thumbnail for any Wikipedia slug.
func FetchPageSummary(slug string) (*WikiPageSummary, error) {
	body, err := getSummaryJSON(slug)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Extract     string `json:"extract"`
		ContentURLs struct {
			Desktop struct {
				Page string `json:"page"`
			} `json:"desktop"`
		} `json:"content_urls"`
		Thumbnail struct {
			Source string `json:"source"`
		} `json:"thumbnail"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("wikimedia: unmarshal response: %w", err)
	}

	return &WikiPageSummary{
		Slug:         slug,
		Extract:      payload.Extract,
		PageURL:      payload.ContentURLs.Desktop.Page,
		ThumbnailURL: payload.Thumbnail.Source,
		FetchedAt:    time.Now().UTC(),
	}, nil
}

// Fetch retrieves the Wikipedia page summary for a national football team.
func Fetch(countrySlug string) (*WikiTeamMeta, error) {
	summary, err := FetchPageSummary(countrySlug)
	if err != nil {
		return nil, err
	}
	return &WikiTeamMeta{
		CountrySlug: countrySlug,
		Extract:     summary.Extract,
		PageURL:     summary.PageURL,
		FetchedAt:   summary.FetchedAt,
	}, nil
}

func getSummaryJSON(slug string) ([]byte, error) {
	url := fmt.Sprintf(
		"https://en.wikipedia.org/api/rest_v1/page/summary/%s",
		slug,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("wikimedia: creating request: %w", err)
	}
	req.Header.Set("User-Agent", "wc-system/1.0 (quantitative football prediction; contact via GitHub)")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wikimedia: HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikimedia: unexpected status %d for slug %s", resp.StatusCode, slug)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("wikimedia: reading body: %w", err)
	}
	return body, nil
}
