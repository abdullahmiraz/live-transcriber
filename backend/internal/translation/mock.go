package translation

import (
	"context"
	"fmt"
	"strings"
)

// MockProvider is a dependency-free provider. It does not really translate; it tags the
// text with the target language so the end-to-end captions pipeline is demonstrable
// without any external account (see docs/stt-decision.md).
type MockProvider struct{}

// NewMock returns a mock translation provider.
func NewMock() *MockProvider { return &MockProvider{} }

// Name implements Provider.
func (m *MockProvider) Name() string { return "mock" }

// Translate implements Provider.
func (m *MockProvider) Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil
	}
	if targetLang == "" || targetLang == sourceLang {
		return text, nil
	}
	return fmt.Sprintf("[%s] %s", targetLang, text), nil
}
