// Package translation defines the translation provider interface and adapters.
// Selection happens via the factory at the composition root (see docs/stt-decision.md).
package translation

import "context"

// Provider translates text from a source language to a target language.
// Implementations: mock (default), libretranslate, deepl, etc.
type Provider interface {
	Name() string
	Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error)
}
