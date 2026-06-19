package teamdata

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"wc-system/adapter/fmcsv"
	"wc-system/catalog"
)

const scraperFMCSV = "FM CSV (Manual Import)"

func computeFMCSVStrength(players []fmcsv.PlayerRecord) float64 {
	if len(players) == 0 {
		return 0
	}
	sorted := append([]fmcsv.PlayerRecord(nil), players...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Overall > sorted[j].Overall
	})
	n := min(11, len(sorted))
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += sorted[i].Overall
	}
	return sum / (float64(n) * 100.0)
}

func (s *Store) persistFMCSVPlayers(ctx context.Context, teamID string, players []fmcsv.PlayerRecord) {
	if s.pool == nil {
		return
	}
	for _, p := range players {
		attrsJSON, err := json.Marshal(p.Attributes)
		if err != nil {
			attrsJSON = []byte("{}")
		}
		_, err = s.pool.Exec(ctx, `
			INSERT INTO player_stats (team_id, name, age, position, nationality, overall, source, attributes)
			VALUES ($1, $2, $3, $4, $5, $6, 'fmcsv', $7::jsonb)
			ON CONFLICT (team_id, name) DO UPDATE SET
				age = EXCLUDED.age,
				position = EXCLUDED.position,
				nationality = EXCLUDED.nationality,
				overall = EXCLUDED.overall,
				source = EXCLUDED.source,
				attributes = EXCLUDED.attributes,
				imported_at = NOW()
		`, teamID, p.Name, p.Age, p.Position, p.Nationality, p.Overall, string(attrsJSON))
		if err != nil {
			log.Printf("teamdata: fmcsv persist %s/%s: %v", teamID, p.Name, err)
		}
	}
}

func fmcsvPlayersToSquad(players []fmcsv.PlayerRecord) []SquadPlayer {
	sorted := append([]fmcsv.PlayerRecord(nil), players...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Overall > sorted[j].Overall
	})

	out := make([]SquadPlayer, 0, len(sorted))
	for i, p := range sorted {
		prof := int(p.Overall / 5)
		if prof < 1 && p.CA > 0 {
			prof = p.CA / 10
		}
		imp := 0
		if p.PA > 0 {
			imp = p.PA / 10
		}
		fitness := int(p.Overall)
		if fitness < 1 && p.CA > 0 {
			fitness = p.CA / 2
		}

		attrs := p.Attributes
		if attrs == nil {
			attrs = map[string]int{}
		}
		groups := p.AttrGroups
		if groups == nil {
			groups = map[string]float64{}
		}

		out = append(out, SquadPlayer{
			ID:         slugPlayerID(p.Name),
			No:         i + 1,
			Name:       p.Name,
			Pos:        p.Position,
			Age:        p.Age,
			Fitness:    fitness,
			Apps:       p.Apps,
			Goals:      p.Goals,
			Value:      p.Value,
			Prof:       prof,
			Imp:        imp,
			Source:     "fmcsv",
			CA:         p.CA,
			PA:         p.PA,
			RCA:        p.RCA,
			Club:       p.Club,
			Height:     p.Height,
			Attributes: attrs,
			AttrGroups: groups,
		})
	}
	return out
}

func slugPlayerID(name string) string {
	name = strings.Split(name, "(")[0]
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
}

func (s *Store) syncFMCSV(ctx context.Context, teamList []catalog.Team) error {
	dir := strings.TrimSpace(s.fmCSVDir)
	file := strings.TrimSpace(s.fmCSVFile)
	if dir == "" && file == "" {
		return nil
	}

	if dir != "" {
		if _, err := os.Stat(dir); err != nil {
			log.Printf("teamdata: fmcsv dir %q: %v (skipping)", dir, err)
			dir = ""
			if file == "" {
				return nil
			}
		}
	}
	if file != "" {
		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("FM_CSV_FILE %q: %w", file, err)
		}
	}

	var master []fmcsv.PlayerRecord
	if file != "" {
		records, err := fmcsv.Parse(fmcsv.FMCSVConfig{FilePath: file})
		if err != nil {
			return err
		}
		master = records
	}

	loaded := 0
	for _, t := range teamList {
		var players []fmcsv.PlayerRecord

		teamPath := filepath.Join(dir, t.ID+".csv")
		if dir != "" {
			if _, err := os.Stat(teamPath); err == nil {
				records, err := fmcsv.Parse(fmcsv.FMCSVConfig{FilePath: teamPath})
				if err != nil {
					log.Printf("teamdata: fmcsv %s: %v", t.ID, err)
					continue
				}
				players = records
			}
		}

		if len(players) == 0 && len(master) > 0 {
			for _, p := range master {
				if catalog.FMNatMatchesAny(t.ID, p.Nationality) {
					players = append(players, p)
				}
			}
		}

		if len(players) == 0 {
			continue
		}

		loaded++
		s.persistFMCSVPlayers(ctx, t.ID, players)

		s.mu.Lock()
		rec := s.teams[t.ID]
		rec.PlayerStrength = computeFMCSVStrength(players)
		s.teams[t.ID] = rec
		s.squads[t.ID] = mergeFMCSVIntoSquad(s.squads[t.ID], fmcsvPlayersToSquad(players))
		s.squadSources[t.ID] = "fmcsv"
		s.mu.Unlock()
	}

	if loaded == 0 && (dir != "" || file != "") {
		log.Printf("teamdata: fmcsv loaded 0 teams from %q", file)
		return nil
	}
	if loaded > 0 {
		log.Printf("teamdata: fmcsv loaded %d teams from %q (W1 PlayerStrength updated)", loaded, file)
	}
	return nil
}

// mergeFMCSVIntoSquad overlays FM attributes onto an existing FIFA squad when names match,
// otherwise returns the FM-only squad sorted by ability.
func mergeFMCSVIntoSquad(existing []SquadPlayer, fm []SquadPlayer) []SquadPlayer {
	if len(existing) == 0 {
		return fm
	}

	out := make([]SquadPlayer, 0, len(fm))
	for _, p := range fm {
		key := normalizePlayerName(p.Name)
		if fifa, ok := findFIFAMatch(existing, key); ok {
			p.WikiSlug = fifa.WikiSlug
			p.ImageURL = fifa.ImageURL
			p.Role = fifa.Role
			if p.Apps == 0 {
				p.Apps = fifa.Apps
			}
			if p.Goals == 0 {
				p.Goals = fifa.Goals
			}
		}
		out = append(out, p)
	}
	return out
}

func normalizePlayerName(name string) string {
	name = strings.Split(name, "(")[0]
	return strings.ToLower(strings.TrimSpace(name))
}

func findFIFAMatch(existing []SquadPlayer, key string) (SquadPlayer, bool) {
	for _, p := range existing {
		if normalizePlayerName(p.Name) == key {
			return p, true
		}
	}
	return SquadPlayer{}, false
}
