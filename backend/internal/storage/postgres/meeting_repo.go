package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"meetingplatform/internal/meeting"
)

// MeetingRepo implements meeting.Repository on top of PostgreSQL.
type MeetingRepo struct {
	pool *pgxpool.Pool
}

// NewMeetingRepo constructs a meeting repository.
func NewMeetingRepo(pool *pgxpool.Pool) *MeetingRepo {
	return &MeetingRepo{pool: pool}
}

var _ meeting.Repository = (*MeetingRepo)(nil)

// Create inserts a meeting and fills server-generated fields (id, created_at).
func (r *MeetingRepo) Create(ctx context.Context, m *meeting.Meeting) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO meetings (slug, title, host_name, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		m.Slug, m.Title, m.HostName, m.Status,
	).Scan(&m.ID, &m.CreatedAt)
}

// GetBySlug returns a meeting or meeting.ErrNotFound.
func (r *MeetingRepo) GetBySlug(ctx context.Context, slug string) (*meeting.Meeting, error) {
	m := &meeting.Meeting{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, title, host_name, status, created_at, ended_at
		FROM meetings WHERE slug = $1`,
		slug,
	).Scan(&m.ID, &m.Slug, &m.Title, &m.HostName, &m.Status, &m.CreatedAt, &m.EndedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, meeting.ErrNotFound
		}
		return nil, err
	}
	return m, nil
}

// End marks a meeting ended and returns the updated row.
func (r *MeetingRepo) End(ctx context.Context, slug string) (*meeting.Meeting, error) {
	m := &meeting.Meeting{}
	err := r.pool.QueryRow(ctx, `
		UPDATE meetings
		SET status = $1, ended_at = now()
		WHERE slug = $2
		RETURNING id, slug, title, host_name, status, created_at, ended_at`,
		meeting.StatusEnded, slug,
	).Scan(&m.ID, &m.Slug, &m.Title, &m.HostName, &m.Status, &m.CreatedAt, &m.EndedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, meeting.ErrNotFound
		}
		return nil, err
	}
	return m, nil
}
