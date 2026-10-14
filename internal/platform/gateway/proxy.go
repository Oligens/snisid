package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type SmartRouter struct {
	routes map[string]*httputil.ReverseProxy
}

func NewSmartRouter() *SmartRouter {
	return &SmartRouter{
		routes: make(map[string]*httputil.ReverseProxy),
	}
}

func (s *SmartRouter) AddRoute(path string, targetURL string) {
	target, _ := url.Parse(targetURL)
	s.routes[path] = httputil.NewSingleHostReverseProxy(target)
}

func (s *SmartRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for path, proxy := range s.routes {
		if strings.HasPrefix(r.URL.Path, path) {
			logger.Info(r.Context(), "Routing external request",
				zap.String("path", r.URL.Path),
				zap.String("route", path),
			)
			proxy.ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "Not Found: No route defined for this intelligence path", http.StatusNotFound)
}
