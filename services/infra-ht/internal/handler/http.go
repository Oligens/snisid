package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snisid/infra-ht/internal/domain"
	"github.com/snisid/infra-ht/internal/service"
)

type Handler struct{ svc *service.InfraService }

func NewHandler(svc *service.InfraService) *Handler { return &Handler{svc: svc} }
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/health", h.Health)
	r.GET("/clusters", h.Clusters)
	r.POST("/dr/drill", h.DRDrill)
}
func (h *Handler) Health(c *gin.Context) { c.JSON(http.StatusOK, h.svc.GetHealth(c.Request.Context())) }
func (h *Handler) Clusters(c *gin.Context) {
	cls, err := h.svc.GetClusters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cls)
}
func (h *Handler) DRDrill(c *gin.Context) {
	var req struct {
		DrillDate    string `json:"drill_date"`
		Scenario     string `json:"scenario"`
		RTOActualMin int    `json:"rto_actual_min"`
		RPOActualMin int    `json:"rpo_actual_min"`
		Success      bool   `json:"success"`
		Notes        string `json:"notes,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d := timeParse("2006-01-02", req.DrillDate)
	drill := domain.DRDrill{
		DrillDate: d, Scenario: req.Scenario,
		RTOActualMin: intPtr(req.RTOActualMin), RPOActualMin: intPtr(req.RPOActualMin),
		Success: boolPtr(req.Success), Notes: strPtr(req.Notes),
	}
	result, err := h.svc.RecordDRDrill(c.Request.Context(), drill)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}
func timeParse(f, v string) time.Time { t, _ := time.Parse(f, v); return t }
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
func intPtr(i int) *int    { return &i }
func boolPtr(b bool) *bool { return &b }
