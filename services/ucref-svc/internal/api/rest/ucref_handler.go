package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/ucref-svc/internal/domain"
	"github.com/snisid/platform/services/ucref-svc/internal/service"
)

type UCREFHandler struct {
	svc *service.UCREFService
	log *zap.Logger
}

func NewUCREFHandler(svc *service.UCREFService, log *zap.Logger) *UCREFHandler {
	return &UCREFHandler{svc: svc, log: log}
}

func (h *UCREFHandler) SubmitSTR(c *gin.Context) {
	var req domain.SubmitSTRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.svc.SubmitSTR(&req)
	if err != nil {
		h.log.Error("submit STR failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"str_id":          report.StrID,
		"national_str_id": report.NationalStrID,
		"status":          report.Status,
	})
}

func (h *UCREFHandler) GetSTRDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid str id"})
		return
	}

	report, err := h.svc.GetSTRDetail(id)
	if err != nil {
		h.log.Error("get STR detail failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "STR not found"})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *UCREFHandler) GetFinancialProfile(c *gin.Context) {
	id, err := uuid.Parse(c.Param("person_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person_id"})
		return
	}

	profile, err := h.svc.GetFinancialProfile(id)
	if err != nil {
		h.log.Error("get financial profile failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *UCREFHandler) RecordMonCashPattern(c *gin.Context) {
	var pattern domain.MonCashPattern
	if err := c.ShouldBindJSON(&pattern); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.RecordMonCashPattern(&pattern); err != nil {
		h.log.Error("record MonCash pattern failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"pattern_id":   pattern.PatternID,
		"str_id":       pattern.STRID,
		"phone_number": pattern.PhoneNumber,
	})
}

func (h *UCREFHandler) GetUnanalyzedSTRs(c *gin.Context) {
	reports, err := h.svc.GetUnanalyzedSTRs()
	if err != nil {
		h.log.Error("get unanalyzed STRs failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reports": reports})
}

func (h *UCREFHandler) DisseminateSTR(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid STR id"})
		return
	}

	var req domain.DisseminateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.DisseminateSTR(id, &req); err != nil {
		h.log.Error("disseminate STR failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "STR disseminated successfully"})
}

func (h *UCREFHandler) GetGangFinances(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid gang id"})
		return
	}

	profiles, err := h.svc.GetGangFinances(id)
	if err != nil {
		h.log.Error("get gang finances failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
}
