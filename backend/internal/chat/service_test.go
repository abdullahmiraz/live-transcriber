package chat

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"meetingplatform/internal/pubsub"
)

type fakeRepo struct {
	saved []*Message
}

func (f *fakeRepo) Save(_ context.Context, m *Message) error {
	m.ID = "msg-id"
	m.CreatedAt = time.Now().UTC()
	f.saved = append(f.saved, m)
	return nil
}

func (f *fakeRepo) ListByMeeting(_ context.Context, _ string, _ int, _ *time.Time) ([]Message, error) {
	out := make([]Message, 0, len(f.saved))
	for _, m := range f.saved {
		out = append(out, *m)
	}
	return out, nil
}

func newTestService() (*Service, *fakeRepo, pubsub.Broker) {
	repo := &fakeRepo{}
	broker := pubsub.NewMemory()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(repo, broker, logger), repo, broker
}

func TestSendPersistsAndPublishes(t *testing.T) {
	svc, repo, broker := newTestService()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := broker.PSubscribe(ctx, "room:*")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	m, err := svc.Send(ctx, SendInput{
		MeetingID:  "mtg-1",
		Slug:       "abc-defg-hij",
		SenderID:   "p1",
		SenderName: "Alice",
		Content:    "  hello team  ",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if m.Content != "hello team" {
		t.Errorf("expected trimmed content, got %q", m.Content)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 persisted message, got %d", len(repo.saved))
	}

	select {
	case msg := <-ch:
		if msg.Channel != "room:abc-defg-hij" {
			t.Errorf("unexpected channel %q", msg.Channel)
		}
		var env struct {
			Type    string  `json:"type"`
			Payload Message `json:"payload"`
		}
		if err := json.Unmarshal(msg.Payload, &env); err != nil {
			t.Fatalf("unmarshal published: %v", err)
		}
		if env.Type != "chat.new" || env.Payload.Content != "hello team" {
			t.Errorf("unexpected published envelope: %+v", env)
		}
	case <-time.After(time.Second):
		t.Fatal("expected a published message, got none")
	}
}

func TestSendValidation(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	if _, err := svc.Send(ctx, SendInput{Content: "   "}); err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
	long := strings.Repeat("x", MaxContentLen+1)
	if _, err := svc.Send(ctx, SendInput{Content: long}); err != ErrContentTooLong {
		t.Errorf("expected ErrContentTooLong, got %v", err)
	}
}
