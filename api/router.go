package api

import (
	"net/http"

	"wc-system/api/handlers"
	"wc-system/config"
	"wc-system/service/teamdata"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// NewRouter constructs and returns the Gin engine with all routes registered.
func NewRouter(pool *pgxpool.Pool, cfg *config.Config, store *teamdata.Store) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.GET("/teams", handlers.TeamsHandler(store))
		api.GET("/teams/:id/squad", handlers.SquadHandler(store))
		api.POST("/predict", handlers.PredictHandler(store))
		api.POST("/predict/monte-carlo", handlers.MonteCarloHandler(store))
		api.POST("/narrative", handlers.NarrativeHandlerFunc(cfg))
		api.GET("/signals", handlers.SignalsHandler(pool, store))
		api.GET("/matches", handlers.MatchesHandler(pool))
		api.GET("/pipeline/status", handlers.PipelineStatusHandler(store))
		api.POST("/pipeline/sync", handlers.PipelineSyncHandler(store))
		api.POST("/pipeline/sync/fbref", handlers.PipelineFBrefSyncHandler(store))
		api.POST("/pipeline/sync/odds", handlers.PipelineOddsSyncHandler(store))
		api.POST("/pipeline/calibrate", handlers.CalibrateHandler(pool, store))
		api.GET("/pipeline/targets", handlers.PipelineTargetsHandler())
		api.GET("/admin/status", handlers.AdminStatusHandler(pool, store))
		api.GET("/wiki/thumbnail/:slug", handlers.WikiThumbnailHandler)
	}

	return r
}
