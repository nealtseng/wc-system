package handlers

import (
	"net/http"
	"time"

	"wc-system/adapter/worldcup26"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type matchEntry struct {
	ID          int64  `json:"id"`
	WCMatchID   string `json:"wc_match_id"`
	HomeID      string `json:"home_id"`
	AwayID      string `json:"away_id"`
	HomeISO2    string `json:"home_iso2"`
	AwayISO2    string `json:"away_iso2"`
	HomeName    string `json:"home_name"`
	AwayName    string `json:"away_name"`
	StadiumName string `json:"stadium_name"`
	Stage       string `json:"stage"`
	Matchday    int    `json:"matchday"`
	LocalDate   string `json:"local_date"`
	Kickoff      string  `json:"kickoff"`
	HomeScore    *int    `json:"home_score"`
	AwayScore    *int    `json:"away_score"`
	Finished     bool    `json:"finished"`
	TimeElapsed  string  `json:"time_elapsed"`
	Status       string  `json:"status"`
}

// MatchesHandler handles GET /api/matches.
func MatchesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(), `
			SELECT m.id,
			       COALESCE(m.wc_match_id,''),
			       COALESCE(m.home_id,''), COALESCE(m.away_id,''),
			       COALESCE(th.iso2,''),   COALESCE(ta.iso2,''),
			       COALESCE(m.home_name,''), COALESCE(m.away_name,''),
			       COALESCE(m.stadium,''),  COALESCE(m.stage,''),
			       COALESCE(m.matchday,0),  COALESCE(m.local_date,''),
			       m.kickoff, m.home_score, m.away_score,
			       COALESCE(m.finished, false), COALESCE(m.time_elapsed,'')
			FROM   matches m
			LEFT JOIN teams th ON th.id = m.home_id
			LEFT JOIN teams ta ON ta.id = m.away_id
			ORDER  BY m.kickoff ASC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		results := make([]matchEntry, 0)
		now := time.Now().UTC()
		for rows.Next() {
			var e matchEntry
			var kickoff time.Time
			var finished bool
			var timeElapsed string
			if err := rows.Scan(
				&e.ID, &e.WCMatchID,
				&e.HomeID, &e.AwayID,
				&e.HomeISO2, &e.AwayISO2,
				&e.HomeName, &e.AwayName,
				&e.StadiumName, &e.Stage,
				&e.Matchday, &e.LocalDate,
				&kickoff, &e.HomeScore, &e.AwayScore,
				&finished, &timeElapsed,
			); err != nil {
				continue
			}
			e.Kickoff = kickoff.UTC().Format(time.RFC3339)
			e.TimeElapsed = timeElapsed
			status, inferredFinished := worldcup26.ResolveMatchStatus(finished, timeElapsed, kickoff.UTC(), now)
			e.Status = status
			e.Finished = finished || inferredFinished
			results = append(results, e)
		}
		c.JSON(http.StatusOK, results)
	}
}
