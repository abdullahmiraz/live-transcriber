package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"meetingplatform/internal/chat"
	"meetingplatform/internal/meeting"
	"meetingplatform/internal/observability"
	"meetingplatform/internal/platform"
)

// --- Health ---

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	platform.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleReadyz(d deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if d.ready != nil {
			if err := d.ready(ctx); err != nil {
				platform.WriteJSON(w, http.StatusServiceUnavailable,
					map[string]string{"status": "unavailable"})
				return
			}
		}
		platform.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

// --- Meetings ---

type createMeetingRequest struct {
	Title    string `json:"title"`
	HostName string `json:"host_name"`
}

type meetingResponse struct {
	*meeting.Meeting
	JoinURL string `json:"join_url"`
}

func toResponse(m *meeting.Meeting) meetingResponse {
	return meetingResponse{Meeting: m, JoinURL: "/m/" + m.Slug}
}

func handleCreateMeeting(d deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createMeetingRequest
		if err := platform.DecodeJSON(r, &req); err != nil {
			platform.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
			return
		}
		m, err := d.meetings.Create(r.Context(), meeting.CreateInput{
			Title:    req.Title,
			HostName: req.HostName,
		})
		if err != nil {
			observability.LoggerFrom(r.Context()).Error("create meeting failed", "error", err)
			platform.WriteError(w, http.StatusInternalServerError, "internal", "could not create meeting")
			return
		}
		observability.LoggerFrom(r.Context()).Info("meeting created",
			"meeting_id", m.ID, "slug", m.Slug)
		platform.WriteJSON(w, http.StatusCreated, toResponse(m))
	}
}

func handleGetMeeting(d deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		m, err := d.meetings.GetBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, meeting.ErrNotFound) {
				platform.WriteError(w, http.StatusNotFound, "not_found", "meeting not found")
				return
			}
			observability.LoggerFrom(r.Context()).Error("get meeting failed", "error", err)
			platform.WriteError(w, http.StatusInternalServerError, "internal", "could not fetch meeting")
			return
		}
		platform.WriteJSON(w, http.StatusOK, toResponse(m))
	}
}

func handleEndMeeting(d deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		m, err := d.meetings.End(r.Context(), slug)
		if err != nil {
			if errors.Is(err, meeting.ErrNotFound) {
				platform.WriteError(w, http.StatusNotFound, "not_found", "meeting not found")
				return
			}
			observability.LoggerFrom(r.Context()).Error("end meeting failed", "error", err)
			platform.WriteError(w, http.StatusInternalServerError, "internal", "could not end meeting")
			return
		}
		platform.WriteJSON(w, http.StatusOK, toResponse(m))
	}
}

// --- Chat ---

type messagesResponse struct {
	Messages []chat.Message `json:"messages"`
}

// handleListMessages returns recent chat history for a meeting (chronological), with
// optional keyset pagination via ?before=<RFC3339> and ?limit=N.
func handleListMessages(d deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		m, err := d.meetings.GetBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, meeting.ErrNotFound) {
				platform.WriteError(w, http.StatusNotFound, "not_found", "meeting not found")
				return
			}
			observability.LoggerFrom(r.Context()).Error("list messages: meeting lookup failed", "error", err)
			platform.WriteError(w, http.StatusInternalServerError, "internal", "could not load meeting")
			return
		}

		limit := 50
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				limit = n
			}
		}
		var before *time.Time
		if v := r.URL.Query().Get("before"); v != "" {
			if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
				before = &t
			} else {
				platform.WriteError(w, http.StatusBadRequest, "bad_request", "before must be RFC3339")
				return
			}
		}

		msgs, err := d.chat.History(r.Context(), m.ID, limit, before)
		if err != nil {
			observability.LoggerFrom(r.Context()).Error("list messages failed", "error", err)
			platform.WriteError(w, http.StatusInternalServerError, "internal", "could not load messages")
			return
		}
		platform.WriteJSON(w, http.StatusOK, messagesResponse{Messages: msgs})
	}
}
