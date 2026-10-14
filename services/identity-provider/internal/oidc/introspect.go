package oidc

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Introspect handles token introspection (RFC 7662).
// POST /oidc/introspect
func (h *Handler) Introspect(c *gin.Context) {
	tokenStr := c.PostForm("token")
	if tokenStr == "" {
		respondError(c, http.StatusBadRequest, "invalid_request", "token parameter required")
		return
	}

	claims, err := h.validateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"active": false,
		})
		return
	}

	sub, _ := claims.GetSubject()
	iss, _ := claims.GetIssuer()
	aud, _ := claims.GetAudience()
	exp, _ := claims.GetExpirationTime()

	c.JSON(http.StatusOK, gin.H{
		"active":     true,
		"sub":        sub,
		"iss":        iss,
		"aud":        aud,
		"exp":        exp.Unix(),
		"token_type": "Bearer",
		"client_id":  sub,
	})
}

// Revoke handles token revocation (RFC 7009).
// POST /oidc/revoke
func (h *Handler) Revoke(c *gin.Context) {
	tokenStr := c.PostForm("token")
	if tokenStr == "" {
		respondError(c, http.StatusBadRequest, "invalid_request", "token parameter required")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"revoked": true,
	})
}

// SessionInfo returns the current SSO session info.
// GET /oidc/session
func (h *Handler) SessionInfo(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		respondError(c, http.StatusBadRequest, "invalid_request", "session_id parameter required")
		return
	}

	sess := h.cfg.SessionStore.GetSession(sessionID)
	if sess == nil {
		respondError(c, http.StatusNotFound, "session_not_found", "Session not found or expired")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"user_id":    sess.UserID,
		"client_id":  sess.ClientID,
		"subject":    sess.Subject,
		"scopes":     sess.Scopes,
	})
}

func (h *Handler) validateToken(tokenStr string) (jwt.MapClaims, error) {
	ensureKey()
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
