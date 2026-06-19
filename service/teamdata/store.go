// Package teamdata synchronises external adapters and serves live team records.
package teamdata

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"wc-system/adapter/fifa"
	"wc-system/adapter/fmscout"
	"wc-system/adapter/kaggle"
	"wc-system/adapter/wikimedia"
	"wc-system/adapter/worldbank"
	"wc-system/adapter/worldcup26"
	"wc-system/catalog"

	"github.com/jackc/pgx/v5/pgxpool"
)

const scraperWorldBank = "World Bank (GDP)"
const scraperWikimedia = "Wikimedia (Squad Meta)"
const scraperKaggle = "Kaggle Hist. Results"
const scraperFIFA = "Kaggle/FIFA Players"
const scraperFBref = "FBref (xG)"
const scraperFMScout = "FMScout (Player Attrs)"
const scraperWorldCup26 = "WorldCup2026 (Fixtures)"
const scraperTheOdds = "The Odds API"

const (
	fmscoutDelayMin = 3 * time.Second
	fmscoutDelayMax = 8 * time.Second
)

// SquadPlayer is one national-team roster entry shown in the squad API.
type SquadPlayer struct {
	ID         string             `json:"id"`
	No         int                `json:"no"`
	Name       string             `json:"name"`
	WikiSlug   string             `json:"wiki_slug"`
	Pos        string             `json:"pos"`
	Age        int                `json:"age"`
	Fitness    int                `json:"fitness"`
	Apps       int                `json:"apps"`
	Goals      int                `json:"goals"`
	Value      string             `json:"value"`
	Prof       int                `json:"prof"`
	Imp        int                `json:"imp"`
	Role       string             `json:"role"`
	OffPitch   float64            `json:"off_pitch"`
	ImageURL   string             `json:"image_url,omitempty"`
	Source     string             `json:"source,omitempty"`
	CA         int                `json:"ca,omitempty"`
	PA         int                `json:"pa,omitempty"`
	RCA        int                `json:"rca,omitempty"`
	Club       string             `json:"club,omitempty"`
	Height     int                `json:"height,omitempty"`
	Attributes map[string]int     `json:"attributes,omitempty"`
	AttrGroups map[string]float64 `json:"attr_groups,omitempty"`
}

// Record is the merged live view of one national team.
type Record struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	ELO             float64 `json:"elo"`
	GDPPerCapita    float64 `json:"gdp_per_capita"`
	GDPYear         int     `json:"gdp_year,omitempty"`
	WikiExtract     string  `json:"wiki_extract,omitempty"`
	WikiURL         string  `json:"wiki_url,omitempty"`
	AvgGoalsFor     float64 `json:"avg_goals_for"`
	WinRate         float64 `json:"win_rate"`
	Momentum        float64 `json:"momentum"`
	AvgXGFor        float64   `json:"avg_xg_for"`
	AvgXGSource     string    `json:"avg_xg_source,omitempty"`
	AvgXGMatchCount int       `json:"avg_xg_match_count,omitempty"`
	AvgXGUpdatedAt  time.Time `json:"avg_xg_updated_at,omitempty"`
	PlayerStrength  float64 `json:"player_strength"`
	MatchesPlayed   int     `json:"matches_played"`
	NarrativeWeight float64 `json:"narrative_weight"`
	Group           string  `json:"wc_group"`
	ISO2            string  `json:"iso2"`
}

// ScraperState tracks one adapter's last sync outcome.
type ScraperState struct {
	Name      string
	Status    string // "ok" | "degraded" | "offline"
	LastFetch time.Time
	Message   string
}

// Store holds synced team data and scraper health.
type Store struct {
	mu         sync.RWMutex
	teams      map[string]Record
	squads        map[string][]SquadPlayer
	squadSources  map[string]string
	scrapers   map[string]ScraperState
	lastSync   time.Time
	fifaCSVURL string
	fmCSVDir   string
	fmCSVFile  string
	oddsAPIKey string
	pool       *pgxpool.Pool
	weights    ModelWeights
	latestOdds map[string]OddsSnapshot
	fbrefXGCsv string
	fbrefSyncMu sync.Mutex
}

// NewStore creates an empty store. fifaCSVURL may be empty to use the default mirror.
// fmCSVDir holds per-team exports ({TEAM_ID}.csv); fmCSVFile is an optional master export.
// fbrefXGCsv is an optional manual xG seed CSV (defaults to data/fbref_xg.csv when empty).
func NewStore(pool *pgxpool.Pool, fifaCSVURL, fmCSVDir, fmCSVFile, oddsAPIKey, fbrefXGCsv string) *Store {
	return &Store{
		teams:      make(map[string]Record),
		squads:       make(map[string][]SquadPlayer),
		squadSources: make(map[string]string),
		scrapers:   make(map[string]ScraperState),
		fifaCSVURL: fifaCSVURL,
		fmCSVDir:   fmCSVDir,
		fmCSVFile:  fmCSVFile,
		oddsAPIKey: oddsAPIKey,
		fbrefXGCsv: fbrefXGCsv,
		pool:       pool,
		weights:    defaultWeights(),
		latestOdds: make(map[string]OddsSnapshot),
	}
}

// SyncAll refreshes World Bank, Wikimedia, and Kaggle data for all catalog teams.
// FBref xG is not scraped here — use SyncFBref or POST /api/pipeline/sync/fbref.
func (s *Store) SyncAll(ctx context.Context) error {
	teamList := catalog.All()
	kaggleNames := catalog.KaggleNames()
	nameList := make([]string, 0, len(teamList))
	for _, t := range teamList {
		nameList = append(nameList, t.KaggleName)
	}

	s.mu.Lock()
	for _, t := range teamList {
		prev := s.teams[t.ID]
		s.teams[t.ID] = Record{
			ID:              t.ID,
			Name:            t.Name,
			ELO:             1500,
			AvgGoalsFor:     1.2,
			Group:           t.Group,
			ISO2:            t.ISO2,
			AvgXGFor:        prev.AvgXGFor,
			AvgXGSource:     prev.AvgXGSource,
			AvgXGMatchCount: prev.AvgXGMatchCount,
			AvgXGUpdatedAt:  prev.AvgXGUpdatedAt,
		}
	}
	s.mu.Unlock()

	var kaggleErr error
	s.syncScraper(scraperKaggle, func() error {
		matches, err := kaggle.FetchResults(kaggleNames)
		if err != nil {
			kaggleErr = err
			return err
		}
		stats := kaggle.ComputeStats(matches, nameList)
		s.mu.Lock()
		for _, t := range teamList {
			rec := s.teams[t.ID]
			if st, ok := stats[t.KaggleName]; ok {
				rec.ELO = st.ELO
				rec.AvgGoalsFor = st.AvgGoalsFor
				rec.WinRate = st.WinRate
				rec.Momentum = st.Momentum
				rec.MatchesPlayed = st.MatchesPlayed
			}
			s.teams[t.ID] = rec
		}
		s.mu.Unlock()
		return nil
	})

	wbFails := 0
	s.syncScraper(scraperWorldBank, func() error {
		for _, t := range teamList {
			result, err := worldbank.Fetch(t.ISO2)
			if err != nil {
				wbFails++
				log.Printf("teamdata: worldbank %s: %v", t.ID, err)
				continue
			}
			s.mu.Lock()
			rec := s.teams[t.ID]
			rec.GDPPerCapita = result.GDPPerCapita
			rec.GDPYear = result.Year
			s.teams[t.ID] = rec
			s.mu.Unlock()
			time.Sleep(200 * time.Millisecond) // respect World Bank rate limits
		}
		if wbFails == len(teamList) {
			return fmt.Errorf("all %d World Bank fetches failed", wbFails)
		}
		if wbFails > 0 {
			return fmt.Errorf("%d/%d World Bank fetches failed", wbFails, len(teamList))
		}
		return nil
	})

	wikiFails := 0
	s.syncScraper(scraperWikimedia, func() error {
		for _, t := range teamList {
			meta, err := wikimedia.Fetch(t.WikiSlug)
			if err != nil {
				wikiFails++
				log.Printf("teamdata: wikimedia %s: %v", t.ID, err)
				continue
			}
			s.mu.Lock()
			rec := s.teams[t.ID]
			rec.WikiExtract = meta.Extract
			rec.WikiURL = meta.PageURL
			s.teams[t.ID] = rec
			s.mu.Unlock()
			time.Sleep(300 * time.Millisecond)
		}
		if wikiFails == len(teamList) {
			return fmt.Errorf("all %d Wikimedia fetches failed", wikiFails)
		}
		if wikiFails > 0 {
			return fmt.Errorf("%d/%d Wikimedia fetches failed", wikiFails, len(teamList))
		}
		return nil
	})

	var fifaErr error
	s.syncScraper(scraperFIFA, func() error {
		squads, err := fifa.FetchSquads(ctx, s.fifaCSVURL, catalog.FIFACountryByTeamID())
		if err != nil {
			fifaErr = err
			return err
		}
		s.mu.Lock()
		for teamID, players := range squads {
			out := make([]SquadPlayer, 0, len(players))
			for _, p := range players {
				out = append(out, SquadPlayer{
					ID:       p.ID,
					No:       p.No,
					Name:     p.Name,
					WikiSlug: p.WikiSlug,
					Pos:      p.Pos,
					Age:      p.Age,
					Fitness:  p.Fitness,
					Value:    p.Value,
					Prof:     p.Prof,
					Imp:      p.Imp,
					Role:     p.Role,
					OffPitch: p.OffPitch,
					ImageURL: p.ImageURL,
				})
			}
			s.squads[teamID] = out
			s.squadSources[teamID] = "fifa"
		}
		s.mu.Unlock()
		return nil
	})

	s.syncScraper(scraperFMScout, func() error {
		for _, t := range teamList {
			fmName, ok := catalog.FMScoutName(t.ID)
			if !ok {
				continue
			}
			players, err := fmscout.FetchNationalTeamPlayers(ctx, fmName)
			if err != nil {
				log.Printf("teamdata: fmscout %s: %v", t.ID, err)
				continue
			}
			s.persistFMScoutPlayers(ctx, t.ID, players)
			s.mu.Lock()
			rec := s.teams[t.ID]
			rec.PlayerStrength = computeFMScoutStrength(players)
			s.teams[t.ID] = rec
			s.mu.Unlock()
			time.Sleep(randDuration(fmscoutDelayMin, fmscoutDelayMax))
		}
		return nil
	})

	s.syncScraper(scraperFMCSV, func() error {
		return s.syncFMCSV(ctx, teamList)
	})

	s.mu.Lock()
	for _, t := range teamList {
		rec := s.teams[t.ID]
		if rec.PlayerStrength == 0 {
			rec.PlayerStrength = squadStrengthFromPlayers(s.squads[t.ID])
		}
		s.teams[t.ID] = rec
	}
	s.mu.Unlock()

	if s.pool != nil {
		if err := s.LoadWeights(ctx); err != nil {
			log.Printf("teamdata: loadWeights: %v", err)
		}
	}

	s.syncScraper(scraperWorldCup26, func() error {
		rawTeams, err := worldcup26.FetchTeams(ctx)
		if err != nil {
			return err
		}

		rawStadiums, err := worldcup26.FetchStadiums(ctx)
		if err != nil {
			return err
		}

		rawMatches, err := worldcup26.FetchMatches(ctx)
		if err != nil {
			return err
		}

		stadiumMap := worldcup26.BuildStadiumMap(rawStadiums)
		wcIDToFIFA := worldcup26.BuildWCIDToFIFAMap(rawTeams)

		s.mu.Lock()
		for _, rt := range rawTeams {
			catTeam, ok := catalog.ByID(rt.FIFACode)
			if !ok {
				continue
			}
			rec := s.teams[rt.FIFACode]
			rec.Group = rt.Groups
			rec.ISO2 = catTeam.ISO2
			s.teams[rt.FIFACode] = rec
		}
		s.mu.Unlock()

		if s.pool == nil {
			return nil
		}

		for _, rt := range rawTeams {
			catTeam, ok := catalog.ByID(rt.FIFACode)
			if !ok {
				continue
			}
			s.mu.Lock()
			rec := s.teams[rt.FIFACode]
			if rec.ELO <= 0 {
				rec.ELO = 1500
			}
			elo := rec.ELO
			s.mu.Unlock()
			_, err := s.pool.Exec(ctx, `
				INSERT INTO teams (id, name, elo, wc_group, iso2)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (id) DO UPDATE SET
					name = EXCLUDED.name,
					wc_group = EXCLUDED.wc_group,
					iso2 = EXCLUDED.iso2
			`, rt.FIFACode, catTeam.Name, elo, rt.Groups, catTeam.ISO2)
			if err != nil {
				log.Printf("teamdata: wc26 team upsert %s: %v", rt.FIFACode, err)
			}
		}

		for _, rm := range rawMatches {
			homeID := wcIDToFIFA[rm.HomeTeamID]
			awayID := wcIDToFIFA[rm.AwayTeamID]
			if rm.HomeTeamID == "0" {
				homeID = ""
			}
			if rm.AwayTeamID == "0" {
				awayID = ""
			}
			stadiumName := stadiumMap[rm.StadiumID]
			matchday, _ := strconv.Atoi(rm.Matchday)

			kickoff, err := time.ParseInLocation("01/02/2006 15:04", rm.LocalDate, worldcup26.StadiumLocation(stadiumName))
			if err != nil {
				log.Printf("teamdata: wc26 parse date %q: %v", rm.LocalDate, err)
				kickoff = time.Time{}
			}

			homeName := rm.HomeNameEN
			if homeName == "" {
				if t, ok := catalog.ByID(homeID); ok {
					homeName = t.Name
				} else if rm.HomeLabel != "" {
					homeName = rm.HomeLabel
				}
			}
			awayName := rm.AwayNameEN
			if awayName == "" {
				if t, ok := catalog.ByID(awayID); ok {
					awayName = t.Name
				} else if rm.AwayLabel != "" {
					awayName = rm.AwayLabel
				}
			}

			var homeIDArg, awayIDArg any
			if homeID != "" {
				homeIDArg = homeID
			}
			if awayID != "" {
				awayIDArg = awayID
			}

			_, err = s.pool.Exec(ctx, `
				INSERT INTO matches
					(wc_match_id, home_id, away_id, kickoff, stadium, stage,
					 matchday, local_date, home_name, away_name,
					 home_score, away_score, finished, time_elapsed)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
				ON CONFLICT (wc_match_id) DO UPDATE SET
					home_id = COALESCE(EXCLUDED.home_id, matches.home_id),
					away_id = COALESCE(EXCLUDED.away_id, matches.away_id),
					kickoff = EXCLUDED.kickoff,
					stadium = EXCLUDED.stadium,
					stage = EXCLUDED.stage,
					matchday = EXCLUDED.matchday,
					local_date = EXCLUDED.local_date,
					home_name = EXCLUDED.home_name,
					away_name = EXCLUDED.away_name,
					home_score = EXCLUDED.home_score,
					away_score = EXCLUDED.away_score,
					finished = EXCLUDED.finished,
					time_elapsed = EXCLUDED.time_elapsed
			`, rm.ID, homeIDArg, awayIDArg, kickoff, stadiumName, rm.Type,
				matchday, rm.LocalDate, homeName, awayName,
				scoreArg(worldcup26.ParseScore(rm.HomeScore)),
				scoreArg(worldcup26.ParseScore(rm.AwayScore)),
				worldcup26.ParseFinished(rm.Finished),
				rm.TimeElapsed)
			if err != nil {
				log.Printf("teamdata: wc26 match insert %s: %v", rm.ID, err)
			}
		}
		return nil
	})

	s.syncScraper(scraperTheOdds, func() error {
		if strings.TrimSpace(s.oddsAPIKey) == "" {
			return fmt.Errorf("THE_ODDS_API_KEY not set")
		}
		if s.pool == nil {
			return nil
		}
		return s.syncTheOdds(ctx)
	})

	s.mu.Lock()
	for id, rec := range s.teams {
		rec.NarrativeWeight = narrativeWeight(rec)
		s.teams[id] = rec
	}
	s.lastSync = time.Now().UTC()

	// Mark partial failures as degraded rather than offline.
	if wbFails > 0 && wbFails < len(teamList) {
		st := s.scrapers[scraperWorldBank]
		st.Status = "degraded"
		st.Message = fmt.Sprintf("%d/%d countries failed", wbFails, len(teamList))
		s.scrapers[scraperWorldBank] = st
	}
	if wikiFails > 0 && wikiFails < len(teamList) {
		st := s.scrapers[scraperWikimedia]
		st.Status = "degraded"
		st.Message = fmt.Sprintf("%d/%d pages failed", wikiFails, len(teamList))
		s.scrapers[scraperWikimedia] = st
	}
	s.mu.Unlock()

	return firstError(kaggleErr, fifaErr)
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) syncScraper(name string, fn func() error) {
	start := time.Now()
	err := fn()
	state := ScraperState{Name: name, LastFetch: start}
	switch {
	case err == nil:
		state.Status = "ok"
	default:
		state.Status = "offline"
		state.Message = err.Error()
	}
	s.mu.Lock()
	s.scrapers[name] = state
	s.mu.Unlock()
}

func narrativeWeight(rec Record) float64 {
	base := rec.WinRate
	if rec.WikiExtract != "" {
		base = base*0.7 + 0.3
	}
	if base > 1 {
		return 1
	}
	if base < 0 {
		return 0
	}
	return base
}

// List returns all team records in catalog order.
func (s *Store) List() []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Record, 0, len(catalog.All()))
	for _, t := range catalog.All() {
		if rec, ok := s.teams[t.ID]; ok {
			out = append(out, rec)
		}
	}
	return out
}

// Get returns one team record.
func (s *Store) Get(id string) (Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.teams[id]
	return rec, ok
}

// Scrapers returns the latest scraper health states.
func (s *Store) Scrapers() []ScraperState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order := []string{scraperWorldBank, scraperWikimedia, scraperKaggle, scraperFIFA, scraperFBref, scraperFMScout, scraperWorldCup26, scraperTheOdds}
	out := make([]ScraperState, 0, len(order))
	for _, name := range order {
		if st, ok := s.scrapers[name]; ok {
			out = append(out, st)
		}
	}
	return out
}

// LastSync returns the timestamp of the most recent sync pass.
func (s *Store) LastSync() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSync
}

// AvgGoalsFor returns expected goals per match for the Poisson engine.
func (s *Store) AvgGoalsFor(id string) float64 {
	v, _ := s.expectedGoals(id)
	return v
}

// W1MicroDelta derives a clipped micro adjustment from squad strength, momentum and win rate.
func (s *Store) W1MicroDelta(homeID, awayID string) float64 {
	home, okH := s.Get(homeID)
	away, okA := s.Get(awayID)
	if !okH || !okA {
		return 0
	}
	w := s.Weights()
	playerDelta := (home.PlayerStrength - away.PlayerStrength) * 0.40
	momentumDelta := (home.Momentum - away.Momentum) * 0.08
	winRateDelta := (home.WinRate - away.WinRate) * 0.05
	delta := playerDelta + momentumDelta + winRateDelta
	return math.Max(-w.DeltaMax, math.Min(w.DeltaMax, delta))
}

// Squad returns the synced roster and data source for a catalog team ID.
func (s *Store) Squad(teamID string) ([]SquadPlayer, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	players, ok := s.squads[teamID]
	if !ok {
		return nil, "", false
	}
	source := s.squadSources[teamID]
	if source == "" {
		source = "fifa"
	}
	out := make([]SquadPlayer, len(players))
	copy(out, players)
	return out, source, true
}
