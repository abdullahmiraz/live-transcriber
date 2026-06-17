package translation

import (
	"context"
	"testing"
)

func TestMockTranslateTagsTarget(t *testing.T) {
	p := NewMock()
	out, err := p.Translate(context.Background(), "hello", "en", "ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "[ru] hello" {
		t.Errorf("expected tagged translation, got %q", out)
	}
}

func TestMockSameLangPassthrough(t *testing.T) {
	p := NewMock()
	out, _ := p.Translate(context.Background(), "hello", "en", "en")
	if out != "hello" {
		t.Errorf("expected passthrough, got %q", out)
	}
}

func TestMockEmptyInput(t *testing.T) {
	p := NewMock()
	out, _ := p.Translate(context.Background(), "   ", "en", "ru")
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}
