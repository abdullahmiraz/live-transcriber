package pubsub

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Redis is a Broker backed by Redis Pub/Sub for multi-instance fan-out. It is also
// compatible with Redis-protocol servers such as Valkey.
type Redis struct {
	client *redis.Client
}

// NewRedis connects to Redis using a URL like redis://host:6379/0 and verifies the
// connection with a ping.
func NewRedis(ctx context.Context, url string) (*Redis, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Redis{client: client}, nil
}

// Name implements Broker.
func (r *Redis) Name() string { return "redis" }

// Publish implements Broker.
func (r *Redis) Publish(ctx context.Context, channel string, payload []byte) error {
	return r.client.Publish(ctx, channel, payload).Err()
}

// PSubscribe implements Broker. The returned channel is closed when ctx is done.
func (r *Redis) PSubscribe(ctx context.Context, pattern string) (<-chan Message, error) {
	ps := r.client.PSubscribe(ctx, pattern)
	out := make(chan Message, 256)

	go func() {
		defer close(out)
		ch := ps.Channel()
		for {
			select {
			case <-ctx.Done():
				_ = ps.Close()
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				select {
				case out <- Message{Channel: msg.Channel, Payload: []byte(msg.Payload)}:
				default:
					// drop on slow consumer
				}
			}
		}
	}()

	return out, nil
}

// Close implements Broker.
func (r *Redis) Close() error { return r.client.Close() }
