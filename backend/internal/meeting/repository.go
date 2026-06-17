package meeting

import "context"

// Repository abstracts persistence. Implemented by storage/postgres; injected at the
// composition root so the domain stays free of infrastructure concerns.
type Repository interface {
	Create(ctx context.Context, m *Meeting) error
	GetBySlug(ctx context.Context, slug string) (*Meeting, error)
	End(ctx context.Context, slug string) (*Meeting, error)
}
