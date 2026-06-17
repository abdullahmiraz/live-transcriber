package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"meetingplatform/internal/chat"
	"meetingplatform/internal/config"
	"meetingplatform/internal/meeting"
	"meetingplatform/internal/observability"
)

// deps bundles everything the transport layer needs, injected from the composition root.
type deps struct {
	cfg       config.Config
	logger    *slog.Logger
	metrics   *observability.Metrics
	meetings  *meeting.Service
	chat      *chat.Service
	ready     func(context.Context) error
	wsHandler http.Handler
}

// Deps is the exported constructor input for the router.
type Deps struct {
	Cfg       config.Config
	Logger    *slog.Logger
	Metrics   *observability.Metrics
	Meetings  *meeting.Service
	Chat      *chat.Service
	Ready     func(context.Context) error
	WSHandler http.Handler
}

// NewRouter builds the fully wired HTTP handler (routes + middleware).
func NewRouter(in Deps) http.Handler {
	d := deps{
		cfg:       in.Cfg,
		logger:    in.Logger,
		metrics:   in.Metrics,
		meetings:  in.Meetings,
		chat:      in.Chat,
		ready:     in.Ready,
		wsHandler: in.WSHandler,
	}

	mux := http.NewServeMux()

	// Health & ops
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", handleReadyz(d))
	if d.metrics != nil {
		mux.Handle("GET /metrics", promhttp.HandlerFor(d.metrics.Registry, promhttp.HandlerOpts{}))
	}

	// Meetings API
	mux.HandleFunc("POST /api/meetings", handleCreateMeeting(d))
	mux.HandleFunc("GET /api/meetings/{slug}", handleGetMeeting(d))
	mux.HandleFunc("POST /api/meetings/{slug}/end", handleEndMeeting(d))

	// Chat history (realtime delivery is over WebSocket)
	if d.chat != nil {
		mux.HandleFunc("GET /api/meetings/{slug}/messages", handleListMessages(d))
	}

	// WebSocket (signaling + realtime events)
	if d.wsHandler != nil {
		mux.Handle("/ws", d.wsHandler)
	}

	return chain(mux,
		recoverer(d),
		requestID(d),
		metricsMW(d),
		cors(d),
	)
}
