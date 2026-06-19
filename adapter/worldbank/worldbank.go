// Package worldbank fetches GDP per capita from the World Bank Open Data API.
// Endpoint: https://api.worldbank.org/v2/country/{iso2}/indicator/NY.GDP.PCAP.CD?format=json&mrv=1
// No API key required.
package worldbank

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// WorldBankResult holds the parsed GDP per capita response for one country.
type WorldBankResult struct {
	CountryCode  string
	GDPPerCapita float64
	Year         int
	FetchedAt    time.Time
}

// Fetch retrieves the most recent GDP per capita for the given ISO-2 country code
// (e.g. "FR", "ES"). Returns an error if the HTTP call fails or the response
// cannot be parsed.
func Fetch(iso2 string) (*WorldBankResult, error) {
	url := fmt.Sprintf(
		"https://api.worldbank.org/v2/country/%s/indicator/NY.GDP.PCAP.CD?format=json&mrv=1",
		iso2,
	)

	resp, err := http.Get(url) //nolint:gosec // URL is constructed from validated iso2
	if err != nil {
		return nil, fmt.Errorf("worldbank: HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("worldbank: unexpected status %d for country %s", resp.StatusCode, iso2)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("worldbank: reading body: %w", err)
	}

	// The World Bank API returns a 2-element JSON array:
	// [0] = metadata object, [1] = array of data points
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("worldbank: unmarshal outer array: %w", err)
	}
	if len(raw) < 2 {
		return nil, fmt.Errorf("worldbank: unexpected response length %d", len(raw))
	}

	var dataPoints []map[string]json.RawMessage
	if err := json.Unmarshal(raw[1], &dataPoints); err != nil {
		return nil, fmt.Errorf("worldbank: unmarshal data array: %w", err)
	}
	if len(dataPoints) == 0 {
		return nil, fmt.Errorf("worldbank: no data points for country %s", iso2)
	}

	first := dataPoints[0]

	// Parse "value" — may be null when data is not yet available
	var gdp float64
	if valRaw, ok := first["value"]; ok {
		var v interface{}
		if err := json.Unmarshal(valRaw, &v); err == nil {
			switch tv := v.(type) {
			case float64:
				gdp = tv
			case nil:
				// value is JSON null — return zero with no error
			}
		}
	}

	// Parse "date" — a 4-character year string
	var year int
	if dateRaw, ok := first["date"]; ok {
		var dateStr string
		if err := json.Unmarshal(dateRaw, &dateStr); err == nil {
			year, _ = strconv.Atoi(dateStr)
		}
	}

	return &WorldBankResult{
		CountryCode:  iso2,
		GDPPerCapita: gdp,
		Year:         year,
		FetchedAt:    time.Now().UTC(),
	}, nil
}
