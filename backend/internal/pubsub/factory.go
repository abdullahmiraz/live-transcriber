package pubsub

import "context"

// New returns a Redis broker when url is non-empty, otherwise an in-memory broker. This
// lets the same code run locally without Redis while using Redis in containers.
func New(ctx context.Context, url string) (Broker, error) {
	if url == "" {
		return NewMemory(), nil
	}
	return NewRedis(ctx, url)
}
