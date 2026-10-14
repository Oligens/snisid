package tracing

import (
	"net/http"
)

// Transport wraps an http.RoundTripper to automatically inject the Correlation ID
// from the request context into the outbound HTTP headers.
type Transport struct {
	Base http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Extract the correlation ID from the context
	corrID := ExtractCorrelationID(req.Context())

	// Create a clone of the request to avoid modifying the original
	clone := req.Clone(req.Context())

	// Inject the header
	clone.Header.Set(CorrelationHeader, corrID)

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	return base.RoundTrip(clone)
}

// NewHTTPClient creates an http.Client equipped with the correlation ID transport wrapper.
func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: &Transport{
			Base: http.DefaultTransport,
		},
	}
}
