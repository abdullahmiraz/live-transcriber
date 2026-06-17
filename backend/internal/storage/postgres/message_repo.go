package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"meetingplatform/internal/chat"
)

// MessageRepo implements chat.Repository on top of PostgreSQL.
type MessageRepo struct {
	pool *pgxpool.Pool
}

// NewMessageRepo constructs a message repository.
func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

var _ chat.Repository = (*MessageRepo)(nil)

// Save inserts a message and fills server-generated fields.
func (r *MessageRepo) Save(ctx context.Context, m *chat.Message) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO messages (meeting_id, sender_id, sender_name, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		m.MeetingID, m.SenderID, m.SenderName, m.Content,
	).Scan(&m.ID, &m.CreatedAt)
}

// ListByMeeting returns messages oldest→newest. It fetches the newest page (optionally
// older than before) using the descending index, then reverses for chronological display.
func (r *MessageRepo) ListByMeeting(ctx context.Context, meetingID string, limit int, before *time.Time) ([]chat.Message, error) {
	const base = `
		SELECT id, meeting_id, sender_id, sender_name, content, created_at
		FROM messages
		WHERE meeting_id = $1`

	var (
		query string
		args  []any
	)
	if before != nil {
		query = base + ` AND created_at < $2 ORDER BY created_at DESC, id DESC LIMIT $3`
		args = []any{meetingID, *before, limit}
	} else {
		query = base + ` ORDER BY created_at DESC, id DESC LIMIT $2`
		args = []any{meetingID, limit}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]chat.Message, 0, limit)
	for rows.Next() {
		var m chat.Message
		if err := rows.Scan(&m.ID, &m.MeetingID, &m.SenderID, &m.SenderName, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// reverse to chronological (oldest → newest)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
