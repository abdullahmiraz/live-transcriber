package meeting

import "context"

// Repository abstracts persistence. Implemented by storage/postgres; injected at the
// composition root so the domain stays free of infrastructure concerns.
type Repository interface {
	Create(ctx context.Context, m *Meeting) error
	GetBySlug(ctx context.Context, slug string) (*Meeting, error)
	End(ctx context.Context, slug string) (*Meeting, error)
	// DeleteBySlug removes the meeting and all related data via FK cascades.
	// Used for automatic cleanup when a room stays empty.
	DeleteBySlug(ctx context.Context, slug string) error
}
