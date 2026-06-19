package handlers

import (
	"errors"
	"net/http"

	"wc-system/service/predict"
	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
)

// predictRequest is the JSON body accepted by POST /api/predict.
type predictRequest struct {
	HomeTeamID string `json:"home_team_id" binding:"required"`
	AwayTeamID string `json:"away_team_id" binding:"required"`
	Neutral    *bool  `json:"neutral"`
}

// PredictHandler handles POST /api/predict using live synced team data.
func PredictHandler(store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req predictRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		isNeutral := wcNeutralDefault(req.Neutral)
		out, err := computeModel(store, req.HomeTeamID, req.AwayTeamID, 10000, 42, isNeutral, nil)
		if err != nil {
			if errors.Is(err, predict.ErrTeamNotSynced) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		homeSrc := store.XGSourceFor(req.HomeTeamID)
		awaySrc := store.XGSourceFor(req.AwayTeamID)
		xgSource := homeSrc
		if homeSrc != awaySrc {
			xgSource = homeSrc + "/" + awaySrc
		}

		c.JSON(http.StatusOK, predictResponseJSON(req.HomeTeamID, req.AwayTeamID, out, xgSource))
	}
}
