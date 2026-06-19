package handlers

import (
	"errors"
	"net/http"

	"wc-system/service/predict"
	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
)

type monteCarloRequest struct {
	HomeTeamID string `json:"home_team_id" binding:"required"`
	AwayTeamID string `json:"away_team_id" binding:"required"`
	Iterations int    `json:"iterations"`
	Neutral    *bool  `json:"neutral"`
}

// MonteCarloHandler runs a high-iteration Poisson simulation and returns updated probabilities.
func MonteCarloHandler(store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req monteCarloRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		iterations := req.Iterations
		if iterations <= 0 {
			iterations = 100000
		}
		if iterations > 500000 {
			iterations = 500000
		}

		isNeutral := wcNeutralDefault(req.Neutral)
		out, err := computeModel(store, req.HomeTeamID, req.AwayTeamID, iterations, monteCarloSeed(), isNeutral, nil)
		if err != nil {
			if errors.Is(err, predict.ErrTeamNotSynced) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := predictResponseJSON(req.HomeTeamID, req.AwayTeamID, out, "")
		resp["iterations"] = iterations
		c.JSON(http.StatusOK, resp)
	}
}
