package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep/internal/domain"
	"github.com/snisid/platform/services/sipep/internal/service"
)

type InmateHandler struct {
	inmateService *service.InmateService
}

func NewInmateHandler(is *service.InmateService) *InmateHandler {
	return &InmateHandler{inmateService: is}
}

func (h *InmateHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/intake", h.Intake)
	rg.GET("/inmates/:id", h.GetInmate)
	rg.GET("/inmates/search", h.SearchInmates)
	rg.POST("/release", h.Release)
}

func (h *InmateHandler) Intake(c *gin.Context) {
	var req domain.IntakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inmate, detention, err := h.inmateService.Intake(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"inmate":    inmate,
		"detention": detention,
	})
}

func (h *InmateHandler) GetInmate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inmate id"})
		return
	}

	inmate, err := h.inmateService.GetInmate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inmate)
}

func (h *InmateHandler) SearchInmates(c *gin.Context) {
	query := c.Query("q")
	inmates, err := h.inmateService.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if inmates == nil {
		inmates = []*domain.Inmate{}
	}

	c.JSON(http.StatusOK, gin.H{"results": inmates})
}

func (h *InmateHandler) Release(c *gin.Context) {
	var req struct {
		InmateID     uuid.UUID             `json:"inmate_id" binding:"required"`
		Release      domain.ReleaseRequest `json:"release" binding:"required"`
		AuthorizedBy uuid.UUID             `json:"authorized_by" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	detention, err := h.inmateService.ProcessRelease(req.InmateID, req.Release, req.AuthorizedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"detention": detention})
}
