package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

func ZeroTrustMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. mTLS Verification (Mock: In prod, check r.TLS.PeerCertificates)
		clientCertSubject := r.Header.Get("X-Client-Cert-Subject")
		if clientCertSubject == "" {
			logger.Warn(r.Context(), "mTLS certificate missing", zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "Unauthorized: mTLS required", http.StatusUnauthorized)
			return
		}

		// 2. JWT Verification
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn(r.Context(), "JWT token missing or invalid", zap.String("client", clientCertSubject))
			http.Error(w, "Unauthorized: JWT required", http.StatusUnauthorized)
			return
		}

		// 3. Inject Security Context
		agencyID := "AGENCY-MOCK-123"
		ctx := context.WithValue(r.Context(), "agency_id", agencyID)

		logger.Info(ctx, "Request authenticated via Zero Trust gateway",
			zap.String("agency", agencyID),
			zap.String("subject", clientCertSubject),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
