package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/domain/identity/entity"
	"github.com/snisid/platform/internal/domain/identity/usecase"
)

type HttpHandler struct {
	svc usecase.IdentityService
}

func NewHttpHandler(svc usecase.IdentityService) *HttpHandler {
	return &HttpHandler{svc: svc}
}

func (h *HttpHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/identities", h.Create)
	r.GET("/identities/:id", h.Get)
	r.PUT("/identities/:id", h.Update)
	r.POST("/identities/:id/flag", h.Flag)
	r.GET("/identities/:id/history", h.GetHistory)
}

// @Summary Create an identity
// @Description Create a new citizen identity record
// @Tags Identity
// @Accept json
// @Produce json
// @Param request body entity.Identity true "Identity Payload"
// @Success 201 {object} entity.Identity
// @Router /v1/identities [post]
func (h *HttpHandler) Create(c *gin.Context) {
	var req entity.Identity
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In real app, extract user from context set by Auth middleware
	changedBy := "system"

	res, err := h.svc.CreateIdentity(c.Request.Context(), &req, changedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// @Summary Get Identity
// @Description Retrieve an identity by ID
// @Tags Identity
// @Produce json
// @Param id path string true "Identity ID"
// @Success 200 {object} entity.Identity
// @Router /v1/identities/{id} [get]
func (h *HttpHandler) Get(c *gin.Context) {
	id := c.Param("id")
	res, err := h.svc.GetIdentity(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "identity not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *HttpHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req entity.Identity
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateFn := func(i *entity.Identity) {
		if req.FirstName != "" {
			i.FirstName = req.FirstName
		}
		if req.LastName != "" {
			i.LastName = req.LastName
		}
		if req.Agency != "" {
			i.Agency = req.Agency
		}
	}

	res, err := h.svc.UpdateIdentity(c.Request.Context(), id, updateFn, "user updated details", "system")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *HttpHandler) Flag(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.svc.FlagIdentity(c.Request.Context(), id, req.Reason, "security-admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "flagged"})
}

func (h *HttpHandler) GetHistory(c *gin.Context) {
	id := c.Param("id")
	res, err := h.svc.GetHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
