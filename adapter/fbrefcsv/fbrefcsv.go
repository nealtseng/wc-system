// Package fbrefcsv loads manual xG seed values from a local CSV file.
package fbrefcsv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Row is one team_id → avg_xg_for mapping from data/fbref_xg.csv.
type Row struct {
	TeamID   string
	AvgXGFor float64
}

// Load reads team_id,avg_xg_for rows. Blank lines and a header row are skipped.
func Load(path string) ([]Row, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("fbrefcsv: open %s: %w", path, err)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("fbrefcsv: read %s: %w", path, err)
	}

	out := make([]Row, 0, len(records))
	for i, rec := range records {
		if len(rec) < 2 {
			continue
		}
		teamID := strings.ToUpper(strings.TrimSpace(rec[0]))
		if teamID == "" || teamID == "TEAM_ID" {
			continue
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil || val <= 0 {
			return nil, fmt.Errorf("fbrefcsv: row %d invalid avg_xg_for %q", i+1, rec[1])
		}
		out = append(out, Row{TeamID: teamID, AvgXGFor: val})
	}
	return out, nil
}
