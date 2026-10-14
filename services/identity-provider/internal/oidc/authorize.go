package oidc

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Authorize handles the OAuth2 authorization request.
// GET /oidc/authorize?response_type=code&client_id=...&redirect_uri=...&scope=...&state=...
func (h *Handler) Authorize(c *gin.Context) {
	responseType := c.Query("response_type")
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	scope := c.DefaultQuery("scope", "openid")
	state := c.Query("state")

	if responseType != "code" {
		respondError(c, http.StatusBadRequest, "unsupported_response_type", "Only authorization_code flow is supported")
		return
	}
	if clientID == "" || !h.cfg.ClientManager.ValidateClient(clientID, "") {
		respondError(c, http.StatusBadRequest, "invalid_client", "Unknown or invalid client")
		return
	}
	expectedURI := h.cfg.ClientManager.GetRedirectURI(clientID)
	if redirectURI != expectedURI {
		respondError(c, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
		return
	}

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"login_uri": h.cfg.Issuer + "/login?client_id=" + clientID + "&redirect_uri=" + redirectURI + "&state=" + state,
			"state":     state,
			"client_id": clientID,
			"scope":     scope,
		})
		return
	}

	if !h.cfg.ConsentEngine.HasConsent(userID, clientID, scope) {
		c.JSON(http.StatusOK, gin.H{
			"consent_required": true,
			"client_id":        clientID,
			"scope":            scope,
			"state":            state,
		})
		return
	}

	sessionID := h.cfg.SessionStore.CreateSession(userID, clientID)

	authCode := generateCode()
	c.JSON(http.StatusOK, gin.H{
		"code":       authCode,
		"state":      state,
		"session_id": sessionID,
	})
}
