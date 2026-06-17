// Package pubsub provides a small publish/subscribe abstraction used to distribute
// realtime events (currently chat) across backend instances. It has two interchangeable
// implementations: Redis (multi-instance) and in-memory (single-instance / local dev).
//
// Design: the chat path publishes to channel "room:{slug}"; the WS hub pattern-subscribes
// to "room:*" once and fans messages out to the matching local room. This keeps the WS
// layer free of provider details and prepares horizontal scaling (see docs/architecture.md).
package pubsub

import "context"

// Message is a payload received on a channel.
type Message struct {
	Channel string
	Payload []byte
}

// Broker publishes and pattern-subscribes to channels.
type Broker interface {
	// Publish sends payload to a channel.
	Publish(ctx context.Context, channel string, payload []byte) error
	// PSubscribe subscribes to a channel pattern (supports a trailing "*" wildcard) and
	// returns a receive-only channel of messages. The subscription lives until ctx is done.
	PSubscribe(ctx context.Context, pattern string) (<-chan Message, error)
	// Name returns the implementation name (for logs/metrics).
	Name() string
	// Close releases resources.
	Close() error
}
