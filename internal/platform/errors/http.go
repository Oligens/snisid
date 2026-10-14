package errors

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/platform/logger"
)

// ErrorResponse represents the standardized JSON structure sent to clients.
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		TraceID string `json:"traceId,omitempty"`
	} `json:"error"`
}

// MapToHTTPStatus converts an internal ErrorCode to standard HTTP status codes.
func MapToHTTPStatus(code ErrorCode) int {
	switch code {
	case InvalidArgument:
		return http.StatusBadRequest
	case Unauthenticated:
		return http.StatusUnauthorized
	case PermissionDenied:
		return http.StatusForbidden
	case NotFound:
		return http.StatusNotFound
	case Conflict:
		return http.StatusConflict
	case Unavailable:
		return http.StatusServiceUnavailable
	case Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// RespondWithError securely translates and writes an error to the gin.Context.
// Internal errors are masked to prevent information leakage, while their real causes are logged.
func RespondWithError(c *gin.Context, err error) {
	// Extract TraceID/CorrelationID
	traceID := c.GetString("X-Correlation-ID") // Assuming telemetry middleware set this

	// Default generic fallback
	httpStatus := http.StatusInternalServerError
	resp := ErrorResponse{}
	resp.Error.Code = string(Internal)
	resp.Error.Message = "An unexpected internal server error occurred"
	resp.Error.TraceID = traceID

	if customErr, ok := err.(*Error); ok {
		httpStatus = MapToHTTPStatus(customErr.Code)

		if customErr.Code == Internal {
			// Log the true underlying error and stack trace for internal failures
			logger.Error(c.Request.Context(), "internal server error: "+customErr.Error(), customErr.Err)
			// Ensure we mask the message sent to the client
			resp.Error.Message = "An unexpected internal server error occurred"
		} else {
			// Safe to expose message for non-internal domains (like Validation, NotFound)
			resp.Error.Code = string(customErr.Code)
			resp.Error.Message = customErr.Message
		}
	} else {
		// Log unhandled non-domain errors
		logger.Error(c.Request.Context(), "unhandled error type", err)
	}

	c.JSON(httpStatus, resp)
}

// RespondWithErrorStd securely translates and writes an error to a standard http.ResponseWriter.
func RespondWithErrorStd(w http.ResponseWriter, r *http.Request, err error) {
	traceID := r.Header.Get("X-Correlation-ID")

	httpStatus := http.StatusInternalServerError
	resp := ErrorResponse{}
	resp.Error.Code = string(Internal)
	resp.Error.Message = "An unexpected internal server error occurred"
	resp.Error.TraceID = traceID

	if customErr, ok := err.(*Error); ok {
		httpStatus = MapToHTTPStatus(customErr.Code)
		if customErr.Code == Internal {
			logger.Error(r.Context(), "internal server error: "+customErr.Error(), customErr.Err)
			resp.Error.Message = "An unexpected internal server error occurred"
		} else {
			resp.Error.Code = string(customErr.Code)
			resp.Error.Message = customErr.Message
		}
	} else {
		logger.Error(r.Context(), "unhandled error type", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}
