// Package transcription defines the speech-to-text provider interface and adapters.
// Nothing outside this package depends on a concrete provider; selection happens via
// the factory at the composition root (see docs/stt-decision.md).
package transcription

import "context"

// Result is a single transcription output.
type Result struct {
	Text    string
	Lang    string
	IsFinal bool
}

// Provider turns an audio chunk into text. Implementations: mock (default), whisper,
// deepgram, etc. Keep this interface minimal and provider-agnostic.
type Provider interface {
	// Name returns the provider identifier (for metrics/logs).
	Name() string
	// Transcribe converts an audio chunk (provider-defined encoding) to text.
	Transcribe(ctx context.Context, audio []byte, lang string) (Result, error)
}
