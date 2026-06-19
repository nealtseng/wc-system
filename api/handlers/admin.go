package handlers

import (
	"net/http"
	"time"

	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminStatusHandler returns live infrastructure and pipeline summary for the admin UI.
func AdminStatusHandler(pool *pgxpool.Pool, store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		scrapers := store.Scrapers()
		okCount := 0
		degradedCount := 0
		for _, s := range scrapers {
			switch s.Status {
			case "ok":
				okCount++
			case "degraded":
				degradedCount++
			}
		}

		lastSync := ""
		if t := store.LastSync(); !t.IsZero() {
			lastSync = t.UTC().Format(time.RFC3339)
		}

		var matchesCount, oddsCount, signalsCount int
		if pool != nil {
			ctx := c.Request.Context()
			_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM matches`).Scan(&matchesCount)
			_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM historical_odds`).Scan(&oddsCount)
			_ = pool.QueryRow(ctx, `
				SELECT COUNT(*) FROM matches m
				JOIN historical_odds ho ON ho.match_id = m.id AND ho.source = 'the-odds-api'
				WHERE m.kickoff > NOW() - INTERVAL '3 hours'
			`).Scan(&signalsCount)
		}

		healthy := okCount > 0 && degradedCount == 0
		if degradedCount > 0 && okCount > 0 {
			healthy = true
		}

		c.JSON(http.StatusOK, gin.H{
			"teams_count":    len(store.List()),
			"matches_count":  matchesCount,
			"odds_count":     oddsCount,
			"signals_count":  signalsCount,
			"scrapers_ok":    okCount,
			"scrapers_total": len(scrapers),
			"scrapers_degraded": degradedCount,
			"last_sync":      lastSync,
			"system_healthy": healthy && lastSync != "",
			"scrapers":       scrapers,
		})
	}
}
