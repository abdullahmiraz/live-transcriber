// Package chat is the domain for in-meeting realtime text chat: model, service, and the
// repository interface. PostgreSQL is the source of truth; realtime fan-out goes through
// the pubsub broker. This package imports neither transport nor concrete infrastructure.
package chat

import (
	"errors"
	"time"
)

// MaxContentLen bounds a single message (defensive limit; text-only, no attachments).
const MaxContentLen = 4000

// Validation errors, mapped to client errors at the transport layer.
var (
	ErrEmptyContent   = errors.New("message content is empty")
	ErrContentTooLong = errors.New("message content too long")
)

// Message is a single chat message persisted to PostgreSQL.
type Message struct {
	ID         string    `json:"id"`
	MeetingID  string    `json:"meetingId"`
	SenderID   string    `json:"senderId"`
	SenderName string    `json:"senderName"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

// RoomChannel is the pubsub channel for a meeting's realtime events.
func RoomChannel(slug string) string { return "room:" + slug }
