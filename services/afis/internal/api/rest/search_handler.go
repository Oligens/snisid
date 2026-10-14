package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/services/afis/internal/domain"
	"github.com/snisid/platform/services/afis/internal/service"
)

type SearchHandler struct {
	search *service.SearchService
}

func NewSearchHandler(s *service.SearchService) *SearchHandler {
	return &SearchHandler{search: s}
}

func (h *SearchHandler) SearchTenprint(c *gin.Context) {
	var req domain.EnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.search.SearchTenprint(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":   results,
		"hit_count": len(results),
		"threshold": service.MinMatchScore,
	})
}

func (h *SearchHandler) SearchLatent(c *gin.Context) {
	var req domain.LatentPrint
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.search.SearchLatent(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":   results,
		"hit_count": len(results),
		"threshold": service.MinMatchScore,
	})
}
