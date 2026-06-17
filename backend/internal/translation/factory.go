package translation

import "fmt"

// New returns a provider by name. Add new providers here (libretranslate, deepl, ...)
// and switch via TRANSLATION_PROVIDER — no other code changes required.
func New(name string) (Provider, error) {
	switch name {
	case "", "mock":
		return NewMock(), nil
	// case "libretranslate":
	//     return NewLibreTranslate(...), nil
	default:
		return nil, fmt.Errorf("unknown translation provider %q", name)
	}
}
