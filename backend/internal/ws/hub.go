package ws

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"meetingplatform/internal/chat"
	"meetingplatform/internal/observability"
	"meetingplatform/internal/pubsub"
	"meetingplatform/internal/transcription"
	"meetingplatform/internal/translation"
)

// Hub owns all active rooms. Presence and signaling fan out in-process; chat fans out via
// the pubsub broker (Redis in containers, in-memory locally), so chat works correctly even
// across multiple backend instances (see docs/architecture.md §8).
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room

	stt               transcription.Provider
	tr                translation.Provider
	chat              *chat.Service
	broker            pubsub.Broker
	defaultTargetLang string
	metrics           *observability.Metrics
	logger            *slog.Logger
}

// NewHub constructs a hub with injected providers, chat service, broker, and observability.
func NewHub(
	stt transcription.Provider,
	tr translation.Provider,
	chatSvc *chat.Service,
	broker pubsub.Broker,
	defaultTargetLang string,
	metrics *observability.Metrics,
	logger *slog.Logger,
) *Hub {
	return &Hub{
		rooms:             make(map[string]*Room),
		stt:               stt,
		tr:                tr,
		chat:              chatSvc,
		broker:            broker,
		defaultTargetLang: defaultTargetLang,
		metrics:           metrics,
		logger:            logger,
	}
}

// Run subscribes to the broker and fans published room messages out to local clients. It
// blocks until ctx is done, so callers should run it in a goroutine.
func (h *Hub) Run(ctx context.Context) {
	ch, err := h.broker.PSubscribe(ctx, "room:*")
	if err != nil {
		h.logger.Error("pubsub subscribe failed", "error", err)
		return
	}
	h.logger.Info("hub subscribed to broker", "broker", h.broker.Name())
	for msg := range ch {
		slug := strings.TrimPrefix(msg.Channel, "room:")
		h.mu.RLock()
		room := h.rooms[slug]
		h.mu.RUnlock()
		if room != nil {
			room.broadcast(msg.Payload, "")
		}
	}
}

// getOrCreateRoom returns the room for slug, creating it (with its meeting id) if needed.
func (h *Hub) getOrCreateRoom(slug, meetingID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.rooms[slug]
	if !ok {
		r = newRoom(slug, meetingID, h)
		h.rooms[slug] = r
		if h.metrics != nil {
			h.metrics.MeetingsActive.Set(float64(len(h.rooms)))
		}
	}
	return r
}

// removeRoom deletes an empty room.
func (h *Hub) removeRoom(slug string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.rooms[slug]; ok && r.isEmpty() {
		delete(h.rooms, slug)
		if h.metrics != nil {
			h.metrics.MeetingsActive.Set(float64(len(h.rooms)))
		}
	}
}

// RoomCount returns the number of active rooms (for diagnostics/metrics).
func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}
