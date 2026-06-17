// Package httpapi is the HTTP transport layer: router, handlers, and middleware.
package httpapi

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"meetingplatform/internal/observability"
	"meetingplatform/internal/platform"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Hijack delegates to the underlying ResponseWriter so WebSocket upgrades work even when
// the request passes through this wrapping middleware.
func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := s.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("underlying ResponseWriter does not support hijacking")
}

// chain applies middlewares in order (outermost first).
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// recoverer converts panics into 500s and logs them.
func recoverer(d deps) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					d.logger.Error("panic recovered", "error", rec, "path", r.URL.Path)
					if d.metrics != nil {
						d.metrics.ErrorsTotal.WithLabelValues("http").Inc()
					}
					platform.WriteError(w, http.StatusInternalServerError, "internal", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// requestID injects a request id into context + logger, and echoes it in a header.
func requestID(d deps) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = "r-" + platform.NewID(10)
			}
			ctx := observability.WithRequestID(r.Context(), id)
			ctx = observability.WithLogger(ctx, d.logger.With("request_id", id))
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// metricsMW records request count + duration labeled by matched route.
func metricsMW(d deps) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if d.metrics == nil {
				return
			}
			route := r.Pattern
			if route == "" {
				route = "unmatched"
			}
			status := http.StatusText(rec.status)
			d.metrics.HTTPRequestsTotal.WithLabelValues(route, r.Method, status).Inc()
			d.metrics.HTTPRequestDuration.WithLabelValues(route, r.Method, status).
				Observe(time.Since(start).Seconds())
		})
	}
}

// cors applies a simple allowlist CORS policy.
func cors(d deps) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(d.cfg.CORSOrigins))
	for _, o := range d.cfg.CORSOrigins {
		allowed[strings.TrimSpace(o)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowed["*"] || allowed[origin]) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
