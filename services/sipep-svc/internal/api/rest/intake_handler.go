package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/services/sipep-svc/internal/service"
)

type IntakeHandler struct {
	intakeSvc *service.IntakeService
}

func NewIntakeHandler(intakeSvc *service.IntakeService) *IntakeHandler {
	return &IntakeHandler{intakeSvc: intakeSvc}
}

func (h *IntakeHandler) ProcessIntake(c *gin.Context) {
	var req service.IntakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inmate, detention, err := h.intakeSvc.ProcessIntake(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"inmate":    inmate,
		"detention": detention,
	})
}
