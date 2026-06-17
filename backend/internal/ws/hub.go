package ws

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"meetingplatform/internal/chat"
	"meetingplatform/internal/meeting"
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
	meetings          *meeting.Service
	broker            pubsub.Broker
	defaultTargetLang string
	metrics           *observability.Metrics
	logger            *slog.Logger

	emptyTTL    time.Duration
	emptyTimers map[string]*time.Timer
	ctx         context.Context
}

// NewHub constructs a hub with injected providers, chat service, broker, and observability.
func NewHub(
	stt transcription.Provider,
	tr translation.Provider,
	chatSvc *chat.Service,
	meetingSvc *meeting.Service,
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
		meetings:          meetingSvc,
		broker:            broker,
		defaultTargetLang: defaultTargetLang,
		metrics:           metrics,
		logger:            logger,
		emptyTTL:          10 * time.Minute,
		emptyTimers:       make(map[string]*time.Timer),
	}
}

// Run subscribes to the broker and fans published room messages out to local clients. It
// blocks until ctx is done, so callers should run it in a goroutine.
func (h *Hub) Run(ctx context.Context) {
	h.ctx = ctx
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

	// If the room was scheduled for deletion (empty), cancel that schedule since someone re-joined.
	if t := h.emptyTimers[slug]; t != nil {
		t.Stop()
		delete(h.emptyTimers, slug)
	}

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

		// Schedule meeting deletion if it remains empty.
		if _, scheduled := h.emptyTimers[slug]; !scheduled {
			h.emptyTimers[slug] = time.AfterFunc(h.emptyTTL, func() {
				if h.ctx != nil && h.ctx.Err() != nil {
					return
				}
				ctx := context.Background()
				if h.ctx != nil {
					ctx = h.ctx
				}
				if err := h.meetings.DeleteBySlug(ctx, slug); err != nil && err != meeting.ErrNotFound {
					h.logger.Error("auto-delete meeting failed", "error", err, "meeting_slug", slug)
					return
				}
				h.logger.Info("auto-deleted empty meeting", "meeting_slug", slug, "ttl", h.emptyTTL.String())
				h.mu.Lock()
				delete(h.emptyTimers, slug)
				h.mu.Unlock()
			})
		}
	}
}

// RoomCount returns the number of active rooms (for diagnostics/metrics).
func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}
