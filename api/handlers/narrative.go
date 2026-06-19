package handlers

import (
	"net/http"

	"wc-system/config"
	"wc-system/semantic/llm"

	"github.com/gin-gonic/gin"
)

// narrativeRequest is the JSON body accepted by POST /api/narrative.
type narrativeRequest struct {
	HomeTeam    string  `json:"home_team" binding:"required"`
	AwayTeam    string  `json:"away_team" binding:"required"`
	HomeWinProb float64 `json:"home_win_prob"`
	DrawProb    float64 `json:"draw_prob"`
	AwayWinProb float64 `json:"away_win_prob"`
	HomeELO     float64 `json:"home_elo"`
	AwayELO     float64 `json:"away_elo"`
	HomeGDP     float64 `json:"home_gdp"`
	AwayGDP     float64 `json:"away_gdp"`
	HomeLambda  float64 `json:"home_lambda"`
	AwayLambda  float64 `json:"away_lambda"`
	W1             float64 `json:"w1"`
	W2             float64 `json:"w2"`
	W3             float64 `json:"w3"`
	VenueLabel     string  `json:"venue_label"`
	PoissonFavors  string  `json:"poisson_favors"`
	FinalFavors    string  `json:"final_favors"`
	SignalConflict bool    `json:"signal_conflict"`
}

// NarrativeHandlerFunc returns a handler bound to LLM configuration.
func NarrativeHandlerFunc(cfg *config.Config) gin.HandlerFunc {
	llmCfg := llm.ResolveConfig(cfg.LLMAPIKey, cfg.LLMBaseURL, cfg.LLMModel)

	return func(c *gin.Context) {
		var req narrativeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, _ := llm.GenerateNarrative(c.Request.Context(), llmCfg, llm.NarrativeInput{
			HomeTeam:       req.HomeTeam,
			AwayTeam:       req.AwayTeam,
			HomeWinProb:    req.HomeWinProb,
			DrawProb:       req.DrawProb,
			AwayWinProb:    req.AwayWinProb,
			HomeELO:        req.HomeELO,
			AwayELO:        req.AwayELO,
			HomeGDP:        req.HomeGDP,
			AwayGDP:        req.AwayGDP,
			HomeLambda:     req.HomeLambda,
			AwayLambda:     req.AwayLambda,
			W1:             req.W1,
			W2:             req.W2,
			W3:             req.W3,
			VenueLabel:     req.VenueLabel,
			PoissonFavors:  req.PoissonFavors,
			FinalFavors:    req.FinalFavors,
			SignalConflict: req.SignalConflict,
		})

		c.JSON(http.StatusOK, gin.H{
			"narrative":  result.Narrative,
			"confidence": result.Confidence,
		})
	}
}
