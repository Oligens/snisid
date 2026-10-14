package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/siar/internal/domain"
	"github.com/snisid/platform/services/siar/internal/service"
)

type FirearmHandler struct {
	svc *service.FirearmService
}

func NewFirearmHandler(svc *service.FirearmService) *FirearmHandler {
	return &FirearmHandler{svc: svc}
}

func (h *FirearmHandler) Create(c *gin.Context) {
	var req domain.CreateFirearmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requête invalide: " + err.Error()})
		return
	}
	createdBy := uuid.Nil
	if uid, ok := c.Request.Context().Value(ContextKeyUserID).(uuid.UUID); ok {
		createdBy = uid
	}
	f, err := h.svc.Create(c.Request.Context(), req, createdBy)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, f)
}

func (h *FirearmHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}
	f, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "arme non trouvée"})
		return
	}
	c.JSON(http.StatusOK, f)
}

func (h *FirearmHandler) CheckSerial(c *gin.Context) {
	sn := c.Param("sn")
	f, err := h.svc.FindBySerial(c.Request.Context(), sn)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"found":   false,
			"message": "numéro de série non trouvé dans le registre",
			"serial":  sn,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"found":   true,
		"firearm": f,
	})
}

func (h *FirearmHandler) StatsByType(c *gin.Context) {
	stats, err := h.svc.StatsByType(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erreur statistiques"})
		return
	}
	if stats == nil {
		stats = []domain.FirearmStatsByType{}
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
