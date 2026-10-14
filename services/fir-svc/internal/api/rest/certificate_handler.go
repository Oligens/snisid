package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/fir-svc/internal/domain"
	"github.com/snisid/platform/services/fir-svc/internal/service"
	"net/http"
)

type CertificateHandler struct {
	certSvc *service.CertificateService
}

func NewCertificateHandler(certSvc *service.CertificateService) *CertificateHandler {
	return &CertificateHandler{certSvc: certSvc}
}

type IssueCertificateRequest struct {
	PersonID string `json:"person_id" binding:"required"`
	Purpose  string `json:"purpose" binding:"required"`
	Office   string `json:"office" binding:"required"`
	IssuedBy string `json:"issued_by" binding:"required"`
}

func (h *CertificateHandler) IssueCertificate(c *gin.Context) {
	var req IssueCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	personID, err := uuid.Parse(req.PersonID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	issuedBy, err := uuid.Parse(req.IssuedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	certReq := domain.CertificateRequest{
		PersonID: personID,
		Purpose:  req.Purpose,
		Office:   req.Office,
		IssuedBy: issuedBy,
	}

	cert, err := h.certSvc.IssueCertificate(c.Request.Context(), certReq, issuedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cert)
}

func (h *CertificateHandler) VerifyCertificate(c *gin.Context) {
	certNumber := c.Param("num")
	if certNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Numéro de certificat requis"})
		return
	}

	cert, err := h.certSvc.VerifyCertificate(c.Request.Context(), certNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificat non trouvé"})
		return
	}

	c.JSON(http.StatusOK, cert)
}
