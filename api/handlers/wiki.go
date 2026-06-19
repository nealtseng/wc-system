package handlers

import (
	"net/http"

	"wc-system/adapter/wikimedia"

	"github.com/gin-gonic/gin"
)

// WikiThumbnailHandler handles GET /api/wiki/thumbnail/:slug
func WikiThumbnailHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	summary, err := wikimedia.FetchPageSummary(slug)
	if err != nil {
		if err == wikimedia.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "wikipedia page not found"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"slug":          summary.Slug,
		"thumbnail_url": summary.ThumbnailURL,
		"page_url":      summary.PageURL,
	})
}
