package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/domain/auth/usecase"
)

type HttpHandler struct {
	svc usecase.AuthService
}

func NewHttpHandler(svc usecase.AuthService) *HttpHandler {
	return &HttpHandler{svc: svc}
}

func (h *HttpHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
	r.POST("/logout", h.Logout)
}

func (h *HttpHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Roles    string `json:"roles"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.Register(c.Request.Context(), req.Username, req.Password, req.Roles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "registered"})
}

func (h *HttpHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Device   string `json:"device"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	clientIP := c.ClientIP()

	res, err := h.svc.Login(c.Request.Context(), req.Username, req.Password, clientIP, req.Device)
	if err != nil {
		status := http.StatusUnauthorized
		if err == usecase.ErrAccountLocked {
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *HttpHandler) Refresh(c *gin.Context) {
	var req struct {
		SessionID    string `json:"sessionId"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	clientIP := c.ClientIP()

	res, err := h.svc.Refresh(c.Request.Context(), req.SessionID, req.RefreshToken, clientIP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *HttpHandler) Logout(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	_ = h.svc.Logout(c.Request.Context(), req.SessionID)
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}
