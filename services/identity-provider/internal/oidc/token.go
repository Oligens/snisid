package oidc

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenRequest represents an OAuth2 token request.
type TokenRequest struct {
	GrantType    string `form:"grant_type"`
	Code         string `form:"code"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	RefreshToken string `form:"refresh_token"`
	Scope        string `form:"scope"`
}

// Token handles the OAuth2 token endpoint.
// POST /oidc/token
func (h *Handler) Token(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	switch req.GrantType {
	case "authorization_code":
		h.handleAuthorizationCode(c, req)
	case "refresh_token":
		h.handleRefreshToken(c, req)
	case "client_credentials":
		h.handleClientCredentials(c, req)
	default:
		respondError(c, http.StatusBadRequest, "unsupported_grant_type", "Grant type not supported")
	}
}

func (h *Handler) handleAuthorizationCode(c *gin.Context, req TokenRequest) {
	if !h.cfg.ClientManager.ValidateClient(req.ClientID, req.ClientSecret) {
		respondError(c, http.StatusUnauthorized, "invalid_client", "Client authentication failed")
		return
	}

	now := time.Now()
	accessToken, err := h.signToken(jwt.MapClaims{
		"iss": h.cfg.Issuer,
		"sub": req.ClientID,
		"aud": req.ClientID,
		"exp": now.Add(1 * time.Hour).Unix(),
		"iat": now.Unix(),
		"jti": uuid.New().String(),
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "server_error", "Failed to sign token")
		return
	}

	refreshToken := uuid.New().String()
	idToken, _ := h.signToken(jwt.MapClaims{
		"iss":   h.cfg.Issuer,
		"sub":   req.ClientID,
		"aud":   req.ClientID,
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
		"nonce": randStr(16),
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
		"id_token":      idToken,
	})
}

func (h *Handler) handleRefreshToken(c *gin.Context, req TokenRequest) {
	if !h.cfg.ClientManager.ValidateClient(req.ClientID, req.ClientSecret) {
		respondError(c, http.StatusUnauthorized, "invalid_client", "Client authentication failed")
		return
	}

	now := time.Now()
	accessToken, err := h.signToken(jwt.MapClaims{
		"iss": h.cfg.Issuer,
		"sub": req.ClientID,
		"aud": req.ClientID,
		"exp": now.Add(1 * time.Hour).Unix(),
		"iat": now.Unix(),
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "server_error", "Failed to sign token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": req.RefreshToken,
	})
}

func (h *Handler) handleClientCredentials(c *gin.Context, req TokenRequest) {
	if !h.cfg.ClientManager.ValidateClient(req.ClientID, req.ClientSecret) {
		respondError(c, http.StatusUnauthorized, "invalid_client", "Client authentication failed")
		return
	}

	now := time.Now()
	accessToken, err := h.signToken(jwt.MapClaims{
		"iss":   h.cfg.Issuer,
		"sub":   req.ClientID,
		"aud":   req.ClientID,
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
		"scope": req.Scope,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "server_error", "Failed to sign token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func (h *Handler) signToken(claims jwt.MapClaims) (string, error) {
	ensureKey()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func generateCode() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "snisid_ac_" + base64.RawURLEncoding.EncodeToString(b)
}

func randStr(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}
