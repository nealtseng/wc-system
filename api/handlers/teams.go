package handlers

import (
	"net/http"

	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
)

// TeamsHandler handles GET /api/teams.
func TeamsHandler(store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		records := store.List()
		if len(records) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "team data not synced yet — call POST /api/pipeline/sync",
			})
			return
		}
		c.JSON(http.StatusOK, records)
	}
}
