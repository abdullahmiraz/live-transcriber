package ws

import (
	"log/slog"
	"sync"

	"meetingplatform/internal/observability"
	"meetingplatform/internal/transcription"
	"meetingplatform/internal/translation"
)

// Hub owns all active rooms. For the MVP this is an in-process registry; the scale path
// is to back fan-out with Redis/NATS across instances (see docs/architecture.md §8).
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room

	stt               transcription.Provider
	tr                translation.Provider
	defaultTargetLang string
	metrics           *observability.Metrics
	logger            *slog.Logger
}

// NewHub constructs a hub with injected providers and observability.
func NewHub(
	stt transcription.Provider,
	tr translation.Provider,
	defaultTargetLang string,
	metrics *observability.Metrics,
	logger *slog.Logger,
) *Hub {
	return &Hub{
		rooms:             make(map[string]*Room),
		stt:               stt,
		tr:                tr,
		defaultTargetLang: defaultTargetLang,
		metrics:           metrics,
		logger:            logger,
	}
}

// getOrCreateRoom returns the room for slug, creating it if needed.
func (h *Hub) getOrCreateRoom(slug string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.rooms[slug]
	if !ok {
		r = newRoom(slug, h)
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
