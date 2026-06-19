package handlers

import (
	"net/http"
	"time"

	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
)

type scraperStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LastFetch string `json:"last_fetch"`
	Message   string `json:"message,omitempty"`
}

type pollingSchedule struct {
	Label      string `json:"label"`
	NextAt     string `json:"next_at"`
	Endpoint   string `json:"endpoint"`
	MatchLabel string `json:"match_label,omitempty"`
	WindowKey  string `json:"window_key,omitempty"`
	WCMatchID  string `json:"wc_match_id,omitempty"`
	Synced     bool   `json:"synced,omitempty"`
}

type pipelineStatusResponse struct {
	Scrapers  []scraperStatus   `json:"scrapers"`
	Schedules []pollingSchedule `json:"schedules"`
	UpdatedAt string            `json:"updated_at"`
}

var scraperTargets = map[string]string{
	"World Bank (GDP)":        "api.worldbank.org/v2/country",
	"Wikimedia (Squad Meta)":  "en.wikipedia.org/api/rest_v1",
	"Kaggle Hist. Results":    "github.com/martj42/international_results",
	"Kaggle/FIFA Players":     "github.com/SolideSpoke/sofifa-web-scraper",
	"FBref (xG)":              "https://fbref.com",
	"FMScout (Player Attrs)":  "https://www.fmscout.com",
	"FM CSV (Manual Import)":  "FM26PlayerExport (local CSV)",
	"WorldCup2026 (Fixtures)": "github.com/rezarahiminia/worldcup2026",
	"The Odds API":            "api.the-odds-api.com/v4/sports/soccer_fifa_world_cup",
}

const oddsEndpoint = "The Odds API"

// PipelineStatusHandler handles GET /api/pipeline/status.
func PipelineStatusHandler(store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now().UTC()
		states := store.Scrapers()

		scrapers := make([]scraperStatus, 0, len(states))
		for _, st := range states {
			lastFetch := ""
			if !st.LastFetch.IsZero() {
				lastFetch = st.LastFetch.UTC().Format(time.RFC3339)
			}
			scrapers = append(scrapers, scraperStatus{
				Name:      st.Name,
				Status:    st.Status,
				LastFetch: lastFetch,
				Message:   st.Message,
			})
		}

		if len(scrapers) == 0 {
			for name := range scraperTargets {
				scrapers = append(scrapers, scraperStatus{
					Name:    name,
					Status:  "offline",
					Message: "awaiting first sync",
				})
			}
		}

		schedules := make([]pollingSchedule, 0)
		upcoming, err := store.NextOddsSchedules(c.Request.Context(), 6)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, u := range upcoming {
			schedules = append(schedules, pollingSchedule{
				Label:      u.Label,
				NextAt:     u.NextAt.UTC().Format(time.RFC3339),
				Endpoint:   oddsEndpoint,
				MatchLabel: u.MatchLabel,
				WindowKey:  u.WindowKey,
				WCMatchID:  u.WCMatchID,
				Synced:     u.Synced,
			})
		}

		updatedAt := now.Format(time.RFC3339)
		if last := store.LastSync(); !last.IsZero() {
			updatedAt = last.UTC().Format(time.RFC3339)
		}

		c.JSON(http.StatusOK, pipelineStatusResponse{
			Scrapers:  scrapers,
			Schedules: schedules,
			UpdatedAt: updatedAt,
		})
	}
}

// PipelineSyncHandler handles POST /api/pipeline/sync (full data pipeline).
func PipelineSyncHandler(store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := store.SyncAll(c.Request.Context()); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":  "degraded",
				"message": err.Error(),
				"teams":   len(store.List()),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"teams":  len(store.List()),
		})
	}
}

// PipelineFBrefSyncHandler handles POST /api/pipeline/sync/fbref (FBref xG only).
func PipelineFBrefSyncHandler(store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := store.SyncFBref(c.Request.Context())
		status := "ok"
		if err != nil {
			status = "degraded"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":       status,
			"teams_ok":     result.TeamsOK,
			"teams_total":  result.TeamsTotal,
			"teams_fail":   result.TeamsFail,
			"message":      result.Message,
			"error":        errString(err),
		})
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// PipelineOddsSyncHandler handles POST /api/pipeline/sync/odds (The Odds API only).
func PipelineOddsSyncHandler(store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := store.SyncOdds(c.Request.Context()); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// PipelineTargetsHandler exposes scraper endpoint labels for the frontend UI.
func PipelineTargetsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, scraperTargets)
	}
}
