package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/interop-ht/internal/domain"
	"github.com/snisid/interop-ht/internal/service"
)

type Handler struct{ svc *service.InteropService }

func NewHandler(svc *service.InteropService) *Handler { return &Handler{svc: svc} }
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/exchange", h.Exchange)
	r.POST("/agreements", h.CreateAgreement)
	r.GET("/logs/:agreement_id", h.GetLogs)
}
func (h *Handler) Exchange(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "exchanged"}) }
func (h *Handler) CreateAgreement(c *gin.Context) {
	var r struct {
		ProviderAgencyID string   `json:"provider_agency_id"`
		ConsumerAgencyID string   `json:"consumer_agency_id"`
		ServiceName      string   `json:"service_name"`
		AllowedFields    []string `json:"allowed_fields"`
		LegalBasis       string   `json:"legal_basis"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := domain.DataExchangeAgreement{
		ServiceName:   r.ServiceName,
		AllowedFields: r.AllowedFields,
		LegalBasis:    strPtr(r.LegalBasis),
	}
	if id, err := uuid.Parse(r.ProviderAgencyID); err == nil {
		req.ProviderAgencyID = id
	}
	if id, err := uuid.Parse(r.ConsumerAgencyID); err == nil {
		req.ConsumerAgencyID = id
	}
	result, err := h.svc.CreateAgreement(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}
func (h *Handler) GetLogs(c *gin.Context) {
	logs, err := h.svc.GetLogs(c.Request.Context(), c.Param("agreement_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "logs not found"})
		return
	}
	c.JSON(http.StatusOK, logs)
}
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
