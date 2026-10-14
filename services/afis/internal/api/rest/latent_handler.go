package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/afis/internal/domain"
	"github.com/snisid/platform/services/afis/internal/service"
)

type LatentHandler struct {
	latent *service.LatentService
	search *service.SearchService
}

func NewLatentHandler(l *service.LatentService, s *service.SearchService) *LatentHandler {
	return &LatentHandler{latent: l, search: s}
}

func (h *LatentHandler) SubmitLatent(c *gin.Context) {
	var req struct {
		CaseReference  string                `json:"case_reference" binding:"required"`
		CrimeSceneID   *uuid.UUID            `json:"crime_scene_id,omitempty"`
		LocationDesc   *string               `json:"location_desc,omitempty"`
		DeptCode       *string               `json:"dept_code,omitempty"`
		FoundAt        string                `json:"found_at" binding:"required"`
		ImageRef       string                `json:"image_ref" binding:"required"`
		FingerPosition domain.FingerPosition `json:"finger_position"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	foundAt, err := time.Parse(time.RFC3339, req.FoundAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "found_at invalide, format ISO8601 requis"})
		return
	}

	if req.FingerPosition == "" {
		req.FingerPosition = domain.FingerUnknown
	}

	latent := domain.LatentPrint{
		CaseReference:  req.CaseReference,
		CrimeSceneID:   req.CrimeSceneID,
		LocationDesc:   req.LocationDesc,
		DeptCode:       req.DeptCode,
		FoundAt:        foundAt,
		ImageRef:       req.ImageRef,
		FingerPosition: req.FingerPosition,
	}

	created, err := h.latent.Submit(c.Request.Context(), latent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *LatentHandler) ConfirmMatch(c *gin.Context) {
	latentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID latente invalide"})
		return
	}

	var req struct {
		SubjectID  uuid.UUID `json:"subject_id" binding:"required"`
		Score      float64   `json:"score" binding:"required"`
		ExaminerID uuid.UUID `json:"examiner_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.latent.ConfirmMatch(c.Request.Context(), latentID, req.SubjectID, req.Score, req.ExaminerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *LatentHandler) GetLatent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	latent, err := h.latent.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, latent)
}

func (h *LatentHandler) ListLatents(c *gin.Context) {
	latents, err := h.latent.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"latents": latents})
}
