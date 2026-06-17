package transcription

import (
	"context"
	"fmt"
	"time"
)

// MockProvider is a deterministic, dependency-free provider so the full pipeline runs
// with `docker compose up` and zero external accounts (see docs/stt-decision.md).
//
// In the MVP, the frontend sends the recognized text (e.g. from the browser Web Speech
// API) in the audio payload as UTF-8 bytes; the mock echoes it back as the transcript.
// If the bytes are not valid text, it returns a placeholder sized to the chunk.
type MockProvider struct{}

// NewMock returns a mock STT provider.
func NewMock() *MockProvider { return &MockProvider{} }

// Name implements Provider.
func (m *MockProvider) Name() string { return "mock" }

// Transcribe implements Provider.
func (m *MockProvider) Transcribe(ctx context.Context, audio []byte, lang string) (Result, error) {
	// simulate a small processing delay so latency metrics are meaningful.
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(10 * time.Millisecond):
	}

	text := string(audio)
	if !isPrintable(text) || text == "" {
		text = fmt.Sprintf("[audio chunk %d bytes]", len(audio))
	}
	return Result{Text: text, Lang: lang, IsFinal: true}, nil
}

func isPrintable(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r == '\uFFFD' {
			return false
		}
	}
	return true
}
