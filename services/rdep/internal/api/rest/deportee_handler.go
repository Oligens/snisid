package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/rdep/internal/domain"
	"github.com/snisid/platform/services/rdep/internal/service"
)

type DeporteeHandler struct {
	deporteeSvc *service.DeporteeService
}

func NewDeporteeHandler(ds *service.DeporteeService) *DeporteeHandler {
	return &DeporteeHandler{deporteeSvc: ds}
}

func (h *DeporteeHandler) Intake(c *gin.Context) {
	var req domain.DeporteeIntakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deportee, err := h.deporteeSvc.Intake(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, deportee)
}

func (h *DeporteeHandler) GetDeportee(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID déporté invalide"})
		return
	}

	deportee, err := h.deporteeSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	events, _ := h.deporteeSvc.GetMonitoringEvents(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"deportee":          deportee,
		"monitoring_events": events,
	})
}

func (h *DeporteeHandler) ScreenDeportee(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID déporté invalide"})
		return
	}

	result, err := h.deporteeSvc.Screen(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *DeporteeHandler) AddMonitoringEvent(c *gin.Context) {
	deporteeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID déporté invalide"})
		return
	}

	var req struct {
		EventType   string    `json:"event_type" binding:"required"`
		EventDate   time.Time `json:"event_date" binding:"required"`
		LocationLat *float64  `json:"location_lat,omitempty"`
		LocationLng *float64  `json:"location_lng,omitempty"`
		Notes       *string   `json:"notes,omitempty"`
		ReportedBy  string    `json:"reported_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reportedBy, err := uuid.Parse(req.ReportedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID officier rapporteur invalide"})
		return
	}

	event := domain.MonitoringEvent{
		EventType:   domain.MonitoringEventType(req.EventType),
		EventDate:   req.EventDate,
		LocationLat: req.LocationLat,
		LocationLng: req.LocationLng,
		Notes:       req.Notes,
		ReportedBy:  reportedBy,
	}

	created, err := h.deporteeSvc.AddMonitoringEvent(c.Request.Context(), deporteeID, event)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *DeporteeHandler) ListHighRisk(c *gin.Context) {
	deportees, err := h.deporteeSvc.ListHighRisk(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deportees": deportees})
}

func (h *DeporteeHandler) ListGangAffiliated(c *gin.Context) {
	deportees, err := h.deporteeSvc.ListGangAffiliated(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deportees": deportees})
}

func (h *DeporteeHandler) StatsByCountry(c *gin.Context) {
	stats, err := h.deporteeSvc.StatsByCountry(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
