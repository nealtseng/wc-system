package handlers

import (
	"net/http"
	"time"

	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type signalEntry struct {
	ID           string       `json:"id"`
	HomeTeam     string       `json:"homeTeam"`
	AwayTeam     string       `json:"awayTeam"`
	HomeFlag     string       `json:"homeFlag"`
	AwayFlag     string       `json:"awayFlag"`
	Kickoff      string       `json:"kickoff"`
	BookmarkOdds oddsTriple   `json:"bookmarkOdds"`
	ImpliedProb  probTriple   `json:"impliedProb"`
	PFinal       probTriple   `json:"pFinal"`
	EV           evTriple     `json:"ev"`
	Weights      weightBundle `json:"weights"`
	KellyFraction kellyTriple `json:"kelly_fraction"`
}

type oddsTriple struct {
	Home float64 `json:"home"`
	Draw float64 `json:"draw"`
	Away float64 `json:"away"`
}

type probTriple struct {
	Home float64 `json:"home"`
	Draw float64 `json:"draw"`
	Away float64 `json:"away"`
}

type evTriple struct {
	Home float64 `json:"home"`
	Draw float64 `json:"draw"`
	Away float64 `json:"away"`
}

type kellyTriple struct {
	Home float64 `json:"home"`
	Draw float64 `json:"draw"`
	Away float64 `json:"away"`
}

type weightBundle struct {
	W1   float64 `json:"w1"`
	W2   float64 `json:"w2"`
	W3   float64 `json:"w3"`
	Clip float64 `json:"clip"`
}

// SignalsHandler handles GET /api/signals from synced odds + model probabilities.
func SignalsHandler(pool *pgxpool.Pool, store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if pool == nil {
			c.JSON(http.StatusOK, []signalEntry{})
			return
		}

		rows, err := pool.Query(c.Request.Context(), `
			SELECT m.id::text,
			       COALESCE(m.home_name, ''),
			       COALESCE(m.away_name, ''),
			       COALESCE(m.home_id, ''),
			       COALESCE(m.away_id, ''),
			       COALESCE(th.iso2, ''),
			       COALESCE(ta.iso2, ''),
			       m.kickoff,
			       ho.home_odds, ho.draw_odds, ho.away_odds
			FROM matches m
			JOIN historical_odds ho ON ho.match_id = m.id AND ho.source = 'the-odds-api'
			LEFT JOIN teams th ON th.id = m.home_id
			LEFT JOIN teams ta ON ta.id = m.away_id
			WHERE m.kickoff > NOW() - INTERVAL '6 hours'
			   OR (m.finished = false AND m.kickoff <= NOW() + INTERVAL '90 days')
			ORDER BY m.kickoff ASC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		w := store.Weights()
		results := make([]signalEntry, 0)
		for rows.Next() {
			var id, homeName, awayName, homeID, awayID, homeISO, awayISO string
			var kickoff time.Time
			var homeOdds, drawOdds, awayOdds float64
			if err := rows.Scan(&id, &homeName, &awayName, &homeID, &awayID, &homeISO, &awayISO, &kickoff, &homeOdds, &drawOdds, &awayOdds); err != nil {
				continue
			}

			impHome, impDraw, impAway := removeVig(homeOdds, drawOdds, awayOdds)
			pHome, pDraw, pAway := impHome, impDraw, impAway

			if _, okH := store.Get(homeID); okH {
				if _, okA := store.Get(awayID); okA {
					out, err := computeModel(store, homeID, awayID, 1000, 42, true, nil)
					if err == nil {
						pHome = out.PFinal["home"]
						pDraw = out.PFinal["draw"]
						pAway = out.PFinal["away"]
					}
				}
			}

			kellyScale := w.KellyScale
			results = append(results, signalEntry{
				ID:       id,
				HomeTeam: homeName,
				AwayTeam: awayName,
				HomeFlag: homeISO,
				AwayFlag: awayISO,
				Kickoff:  kickoff.UTC().Format(time.RFC3339),
				BookmarkOdds: oddsTriple{homeOdds, drawOdds, awayOdds},
				ImpliedProb:  probTriple{impHome, impDraw, impAway},
				PFinal:       probTriple{pHome, pDraw, pAway},
				EV: evTriple{
					Home: pHome*homeOdds - 1,
					Draw: pDraw*drawOdds - 1,
					Away: pAway*awayOdds - 1,
				},
				Weights: weightBundle{w.W1, w.W2, w.W3, w.ClipDelta},
				KellyFraction: kellyTriple{
					Home: FractionalKelly(pHome, homeOdds, kellyScale),
					Draw: FractionalKelly(pDraw, drawOdds, kellyScale),
					Away: FractionalKelly(pAway, awayOdds, kellyScale),
				},
			})
		}
		c.JSON(http.StatusOK, results)
	}
}
