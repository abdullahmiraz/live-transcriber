package transcription

import "fmt"

// New returns a provider by name. Add new providers here (whisper, deepgram, ...) and
// switch via the STT_PROVIDER env var — no other code changes required.
func New(name string) (Provider, error) {
	switch name {
	case "", "mock":
		return NewMock(), nil
	// case "whisper":
	//     return NewWhisper(...), nil
	// case "deepgram":
	//     return NewDeepgram(...), nil
	default:
		return nil, fmt.Errorf("unknown STT provider %q", name)
	}
}
