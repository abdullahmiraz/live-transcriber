package transcription

import (
	"context"
	"strings"
	"testing"
)

func TestMockEchoesText(t *testing.T) {
	p := NewMock()
	res, err := p.Transcribe(context.Background(), []byte("hello world"), "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Text != "hello world" {
		t.Errorf("expected echoed text, got %q", res.Text)
	}
	if res.Lang != "en" || !res.IsFinal {
		t.Errorf("unexpected result metadata: %+v", res)
	}
}

func TestMockPlaceholderForBinary(t *testing.T) {
	p := NewMock()
	res, err := p.Transcribe(context.Background(), []byte{0xff, 0xfe, 0x00}, "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(res.Text, "[audio chunk") {
		t.Errorf("expected placeholder, got %q", res.Text)
	}
}
