package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/snisid/governance-svc/internal/domain"
	"github.com/snisid/governance-svc/internal/service"
)

type Handler struct {
	svc *service.GovernanceService
}

func NewHandler(svc *service.GovernanceService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/licenses", h.RegisterLicense)
	r.GET("/licenses", h.ListLicenses)
	r.POST("/policies", h.CreatePolicy)
	r.GET("/policies", h.ListPolicies)
	r.POST("/compliance/check", h.CheckCompliance)
	r.GET("/compliance/report", h.ComplianceReport)
	r.GET("/attribution", h.AttributionReport)
}

func (h *Handler) RegisterLicense(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		Name          string `json:"name"`
		SPDXID        string `json:"spdx_id"`
		LicenseType   string `json:"license_type"`
		Version       string `json:"version"`
		Publisher     string `json:"publisher"`
		IsOsiApproved bool   `json:"is_osi_approved"`
		Text          string `json:"text,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	result, err := h.svc.RegisterLicense(c.Request.Context(),
		req.Name, req.SPDXID, domain.LicenseType(req.LicenseType),
		req.Version, req.Publisher, req.IsOsiApproved, req.Text,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) ListLicenses(c *gin.Context) {
	result, err := h.svc.ListLicenses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) CreatePolicy(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	result, err := h.svc.CreatePolicy(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) ListPolicies(c *gin.Context) {
	result, err := h.svc.ListPolicies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) CheckCompliance(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		LicenseID string `json:"license_id"`
		PolicyID  string `json:"policy_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	licenseID, _ := uuid.Parse(req.LicenseID)
	policyID, _ := uuid.Parse(req.PolicyID)

	result, err := h.svc.CheckCompliance(c.Request.Context(), licenseID, policyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ComplianceReport(c *gin.Context) {
	result, err := h.svc.GenerateComplianceReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) AttributionReport(c *gin.Context) {
	result, err := h.svc.GenerateAttribution(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
