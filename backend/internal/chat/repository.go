package chat

import (
	"context"
	"time"
)

// Repository abstracts message persistence. Implemented by storage/postgres.
type Repository interface {
	// Save inserts a message, filling server-generated fields (id, created_at).
	Save(ctx context.Context, m *Message) error
	// ListByMeeting returns up to limit messages for a meeting in chronological order
	// (oldest → newest). When before is non-nil, only messages strictly older than it are
	// returned (keyset pagination for loading earlier history).
	ListByMeeting(ctx context.Context, meetingID string, limit int, before *time.Time) ([]Message, error)
}
