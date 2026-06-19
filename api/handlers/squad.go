package handlers

import (
	"net/http"

	"wc-system/catalog"
	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
)

type squadResponse struct {
	TeamID  string                 `json:"team_id"`
	Source  string                 `json:"source"`
	Players []teamdata.SquadPlayer `json:"players"`
}

// SquadHandler handles GET /api/teams/:id/squad.
func SquadHandler(store *teamdata.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamID := c.Param("id")
		if _, ok := catalog.ByID(teamID); !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "unknown team id"})
			return
		}

		if store.LastSync().IsZero() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":  "sync_in_progress",
				"error": "資料同步進行中，請稍候再試（或至管線頁面手動觸發 POST /api/pipeline/sync）",
			})
			return
		}

		players, source, ok := store.Squad(teamID)
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":  "fifa_sync_failed",
				"error": "FIFA 名單同步失敗，請至管線頁面重新同步",
			})
			return
		}

		c.JSON(http.StatusOK, squadResponse{
			TeamID:  teamID,
			Source:  source,
			Players: players,
		})
	}
}
