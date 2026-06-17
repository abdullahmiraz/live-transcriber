package meeting

import (
	"context"
	"strings"
	"time"

	"meetingplatform/internal/platform"
)

// Service holds meeting business logic. Transport layers call this; it calls the repo.
type Service struct {
	repo Repository
}

// NewService constructs a meeting service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput is the data needed to create a meeting.
type CreateInput struct {
	Title    string
	HostName string
}

// Create makes a new active meeting with a unique slug.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Meeting, error) {
	m := &Meeting{
		Slug:      platform.NewMeetingSlug(),
		Title:     strings.TrimSpace(in.Title),
		HostName:  strings.TrimSpace(in.HostName),
		Status:    StatusActive,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetBySlug fetches a meeting by its shareable slug.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*Meeting, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// End marks a meeting as ended.
func (s *Service) End(ctx context.Context, slug string) (*Meeting, error) {
	return s.repo.End(ctx, slug)
}
