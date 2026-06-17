package ws

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"meetingplatform/internal/meeting"
	"meetingplatform/internal/platform"
)

// Handler upgrades HTTP requests to WebSocket and joins the requested room.
type Handler struct {
	hub            *Hub
	meetings       *meeting.Service
	allowedOrigins map[string]bool
	upgrader       websocket.Upgrader
}

// NewHandler builds the WS handler. allowedOrigins controls the Origin check; an empty
// set allows all (useful for local dev behind nginx).
func NewHandler(hub *Hub, meetings *meeting.Service, allowedOrigins []string) *Handler {
	set := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		set[strings.TrimSpace(o)] = true
	}
	h := &Handler{hub: hub, meetings: meetings, allowedOrigins: set}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     h.checkOrigin,
	}
	return h
}

func (h *Handler) checkOrigin(r *http.Request) bool {
	if len(h.allowedOrigins) == 0 || h.allowedOrigins["*"] {
		return true
	}
	origin := r.Header.Get("Origin")
	return origin == "" || h.allowedOrigins[origin]
}

// ServeHTTP handles GET /ws?meeting={slug}&name={displayName}.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.URL.Query().Get("meeting"))
	if slug == "" {
		platform.WriteError(w, http.StatusBadRequest, "bad_request", "missing meeting")
		return
	}
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		name = "Guest"
	}

	// Resolve the meeting before upgrading so we can return a proper HTTP error and so the
	// room knows its database id (needed to persist chat messages).
	m, err := h.meetings.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, meeting.ErrNotFound) {
			platform.WriteError(w, http.StatusNotFound, "not_found", "meeting not found")
			return
		}
		platform.WriteError(w, http.StatusInternalServerError, "internal", "could not load meeting")
		return
	}
	if m.Status == meeting.StatusEnded {
		platform.WriteError(w, http.StatusConflict, "conflict", "meeting has ended")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return // upgrader already wrote the error
	}

	room := h.hub.getOrCreateRoom(slug, m.ID)
	client := newClient(platform.NewID(12), name, conn, room, h.hub)

	peers := room.add(client)
	if h.hub.metrics != nil {
		h.hub.metrics.WSConnectionsActive.Inc()
	}

	// Greet the new client with its id and the current roster.
	client.trySend(encode(Envelope{
		Type: TypeWelcome,
		Payload: mustMarshal(WelcomePayload{
			SelfID:       client.id,
			Participants: peers,
		}),
	}))

	// Tell existing peers someone joined.
	room.broadcast(encode(Envelope{
		Type:    TypeParticipantJoined,
		Payload: mustMarshal(PeerInfo{ID: client.id, Name: client.name}),
	}), client.id)

	h.hub.logger.Info("ws client joined",
		"participant_id", client.id, "meeting_id", slug, "name", name)

	go client.writePump()
	go client.readPump()
}
