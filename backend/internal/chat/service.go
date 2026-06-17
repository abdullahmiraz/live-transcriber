package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"meetingplatform/internal/pubsub"
)

// Service holds chat business logic: validate → persist (source of truth) → publish to
// the realtime broker for fan-out to all participants across instances.
type Service struct {
	repo   Repository
	broker pubsub.Broker
	logger *slog.Logger
}

// NewService constructs a chat service.
func NewService(repo Repository, broker pubsub.Broker, logger *slog.Logger) *Service {
	return &Service{repo: repo, broker: broker, logger: logger}
}

// SendInput is the data needed to send a message.
type SendInput struct {
	MeetingID  string
	Slug       string
	SenderID   string
	SenderName string
	Content    string
}

// wireEnvelope mirrors the WebSocket envelope shape so subscribers can forward published
// bytes to browsers verbatim, without this package importing the ws transport package.
type wireEnvelope struct {
	Type    string   `json:"type"`
	Payload *Message `json:"payload"`
}

// Send validates and persists a message, then publishes it to the room channel.
func (s *Service) Send(ctx context.Context, in SendInput) (*Message, error) {
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return nil, ErrEmptyContent
	}
	if utf8.RuneCountInString(content) > MaxContentLen {
		return nil, ErrContentTooLong
	}

	m := &Message{
		MeetingID:  in.MeetingID,
		SenderID:   in.SenderID,
		SenderName: in.SenderName,
		Content:    content,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.repo.Save(ctx, m); err != nil {
		return nil, err
	}

	// Persistence succeeded; publish for realtime delivery. A publish failure is logged
	// but not returned — the message is durably stored and will appear on history load.
	data, err := json.Marshal(wireEnvelope{Type: "chat.new", Payload: m})
	if err == nil {
		if perr := s.broker.Publish(ctx, RoomChannel(in.Slug), data); perr != nil {
			s.logger.Error("chat publish failed", "error", perr, "meeting_id", in.MeetingID)
		}
	}
	return m, nil
}

// History returns recent messages (chronological), optionally older than before.
func (s *Service) History(ctx context.Context, meetingID string, limit int, before *time.Time) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListByMeeting(ctx, meetingID, limit, before)
}
