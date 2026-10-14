package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/snisid/pki-bridge-svc/internal/domain"
	"github.com/snisid/pki-bridge-svc/internal/service"
)

type Handler struct {
	svc *service.PKIBridgeService
}

func NewHandler(svc *service.PKIBridgeService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/foreign-cas", h.RegisterForeignCA)
	r.POST("/cross-certs", h.IssueCrossCert)
	r.GET("/cross-certs/:subject", h.GetCrossCert)
	r.GET("/trust-anchors", h.ListTrustAnchors)
	r.POST("/validate-path", h.ValidatePath)
	r.GET("/bridges", h.ListAgreements)
	r.POST("/bridges", h.CreateAgreement)
}

func (h *Handler) RegisterForeignCA(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		Name         string  `json:"name"`
		Country      string  `json:"country"`
		PublicKeyPEM string  `json:"public_key_pem"`
		CertPolicy   *string `json:"cert_policy,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	ca := domain.ForeignCA{
		Name:         req.Name,
		Country:      req.Country,
		PublicKeyPEM: req.PublicKeyPEM,
		CertPolicy:   req.CertPolicy,
	}

	result, err := h.svc.RegisterForeignCA(c.Request.Context(), ca)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) IssueCrossCert(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		Subject        string `json:"subject"`
		IssuerCAID     string `json:"issuer_ca_id"`
		SerialNumber   string `json:"serial_number"`
		NotBefore      string `json:"not_before"`
		NotAfter       string `json:"not_after"`
		CertificatePEM string `json:"certificate_pem"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	issuerCAID, err := uuid.Parse(req.IssuerCAID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issuer_ca_id"})
		return
	}

	notBefore, err := time.Parse("2006-01-02", req.NotBefore)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid not_before, use YYYY-MM-DD"})
		return
	}

	notAfter, err := time.Parse("2006-01-02", req.NotAfter)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid not_after, use YYYY-MM-DD"})
		return
	}

	result, err := h.svc.IssueCrossCert(c.Request.Context(), req.Subject, issuerCAID, req.SerialNumber, notBefore, notAfter, req.CertificatePEM)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetCrossCert(c *gin.Context) {
	subject := c.Param("subject")
	cert, err := h.svc.GetCrossCert(c.Request.Context(), subject)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cross certificate not found"})
		return
	}
	c.JSON(http.StatusOK, cert)
}

func (h *Handler) ListTrustAnchors(c *gin.Context) {
	anchors, err := h.svc.ListTrustAnchors(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": anchors})
}

func (h *Handler) ValidatePath(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		LeafSubject   string   `json:"leaf_subject"`
		Intermediates []string `json:"intermediates"`
		RootSubject   string   `json:"root_subject"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	result, err := h.svc.ValidatePath(c.Request.Context(), req.LeafSubject, req.Intermediates, req.RootSubject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListAgreements(c *gin.Context) {
	agreements, err := h.svc.ListAgreements(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": agreements})
}

func (h *Handler) CreateAgreement(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		Name      string  `json:"name"`
		PartnerCA string  `json:"partner_ca"`
		PolicyID  string  `json:"policy_id"`
		ExpiresAt *string `json:"expires_at,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	policyID, err := uuid.Parse(req.PolicyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy_id"})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse("2006-01-02", *req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_at"})
			return
		}
		expiresAt = &t
	}

	result, err := h.svc.CreateAgreement(c.Request.Context(), req.Name, req.PartnerCA, policyID, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}
