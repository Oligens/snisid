package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/sipep-svc/internal/domain"
	"github.com/snisid/platform/services/sipep-svc/internal/service"
)

type InmateHandler struct {
	intakeSvc   *service.IntakeService
	releaseSvc  *service.ReleaseService
	transferSvc *service.TransferService
}

func NewInmateHandler(
	intakeSvc *service.IntakeService,
	releaseSvc *service.ReleaseService,
	transferSvc *service.TransferService,
) *InmateHandler {
	return &InmateHandler{
		intakeSvc:   intakeSvc,
		releaseSvc:  releaseSvc,
		transferSvc: transferSvc,
	}
}

func (h *InmateHandler) GetInmate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	inmate, err := h.intakeSvc.GetInmate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Détenu non trouvé"})
		return
	}

	c.JSON(http.StatusOK, inmate)
}

func (h *InmateHandler) SearchInmates(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paramètre de recherche requis"})
		return
	}

	inmates, err := h.intakeSvc.SearchInmates(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": inmates})
}

func (h *InmateHandler) ProcessRelease(c *gin.Context) {
	var req struct {
		InmateID     string `json:"inmate_id" binding:"required"`
		ReleaseType  string `json:"release_type" binding:"required"`
		Authority    string `json:"authority" binding:"required"`
		AuthorizedBy string `json:"authorized_by" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inmateID, err := uuid.Parse(req.InmateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	authorizedBy, err := uuid.Parse(req.AuthorizedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	releaseReq := domain.ReleaseRequest{
		ReleaseType: domain.ReleaseType(req.ReleaseType),
		Authority:   req.Authority,
	}

	detention, err := h.releaseSvc.ProcessRelease(c.Request.Context(), inmateID, releaseReq, authorizedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Libération traitée avec succès",
		"detention": detention,
	})
}

func (h *InmateHandler) ProcessTransfer(c *gin.Context) {
	var req service.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.transferSvc.ProcessTransfer(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Transfert effectué",
		"transfer": transfer,
	})
}

func (h *InmateHandler) GetTransfers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	transfers, err := h.transferSvc.GetTransfers(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transfers": transfers})
}
