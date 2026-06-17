package pubsub

import (
	"context"
	"strings"
	"sync"
)

// Memory is an in-process Broker for single-instance deployments and local dev (no Redis
// required). It supports a trailing "*" wildcard in subscription patterns.
type Memory struct {
	mu   sync.RWMutex
	subs []*memSub
}

type memSub struct {
	pattern string
	ch      chan Message
}

// NewMemory creates an in-memory broker.
func NewMemory() *Memory { return &Memory{} }

// Name implements Broker.
func (m *Memory) Name() string { return "memory" }

// Publish implements Broker.
func (m *Memory) Publish(_ context.Context, channel string, payload []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.subs {
		if matchPattern(s.pattern, channel) {
			select {
			case s.ch <- Message{Channel: channel, Payload: append([]byte(nil), payload...)}:
			default:
				// drop on slow subscriber to protect the publisher
			}
		}
	}
	return nil
}

// PSubscribe implements Broker.
func (m *Memory) PSubscribe(ctx context.Context, pattern string) (<-chan Message, error) {
	sub := &memSub{pattern: pattern, ch: make(chan Message, 256)}
	m.mu.Lock()
	m.subs = append(m.subs, sub)
	m.mu.Unlock()

	go func() {
		<-ctx.Done()
		m.mu.Lock()
		for i, s := range m.subs {
			if s == sub {
				m.subs = append(m.subs[:i], m.subs[i+1:]...)
				break
			}
		}
		m.mu.Unlock()
		close(sub.ch)
	}()

	return sub.ch, nil
}

// Close implements Broker.
func (m *Memory) Close() error { return nil }

// matchPattern supports exact match and a single trailing "*" prefix wildcard.
func matchPattern(pattern, channel string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(channel, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == channel
}
