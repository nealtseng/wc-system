package handlers

import (
	"log"
	"net/http"

	"wc-system/service/calibration"
	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CalibrateHandler handles POST /api/pipeline/calibrate.
func CalibrateHandler(pool *pgxpool.Pool, store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := calibration.Calibrate(c.Request.Context(), pool, store)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		newW := teamdata.ModelWeights{
			W1: result.W1, W2: result.W2, W3: result.W3,
			ClipDelta:  store.Weights().ClipDelta,
			DeltaMax:   store.Weights().DeltaMax,
			KellyScale: store.Weights().KellyScale,
		}
		if saveErr := store.SaveWeights(c.Request.Context(), newW); saveErr != nil {
			log.Printf("calibrate: save weights: %v", saveErr)
		}
		c.JSON(http.StatusOK, gin.H{
			"w1":           result.W1,
			"w2":           result.W2,
			"w3":           result.W3,
			"brier_score":  result.BrierScore,
			"matches_used": result.MatchesUsed,
		})
	}
}
