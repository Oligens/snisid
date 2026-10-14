package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snisid/iam-ht/internal/service"
)

type Handler struct {
	svc *service.IAMService
}

func NewHandler(svc *service.IAMService) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/.well-known/openid-config", h.Discovery)
	r.POST("/authorize", h.Authorize)
	r.POST("/token", h.Token)
	r.POST("/step-up", h.StepUp)
}

func (h *Handler) Discovery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"issuer":                   "https://sso.gouv.ht",
		"authorization_endpoint":   "/api/v1/iam/authorize",
		"token_endpoint":           "/api/v1/iam/token",
		"jwks_uri":                 "/api/v1/iam/.well-known/jwks",
		"response_types_supported": []string{"code", "token"},
		"scopes_supported":         []string{"openid", "profile", "identity"},
	})
}

func (h *Handler) Authorize(c *gin.Context) {
	citizenID := c.Query("citizen_id")
	assurance, err := h.svc.Authorize(c.Request.Context(), citizenID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"assurance_level": assurance.AssuranceLevel, "mfa_enrolled": assurance.MFAEnrolled})
}

func (h *Handler) Token(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"access_token": "mock-token", "token_type": "Bearer", "expires_in": 3600})
}

func (h *Handler) StepUp(c *gin.Context) {
	var req struct {
		CitizenID string `json:"citizen_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.svc.StepUpAssurance(c.Request.Context(), req.CitizenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"assurance_level": result.AssuranceLevel})
}
