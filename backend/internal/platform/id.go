// Package platform holds small shared utilities (ids, http helpers).
package platform

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// avoid ambiguous characters (no 0/o/1/l/i) for readable, shareable slugs.
const slugAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// NewMeetingSlug returns a Meet-style slug like "abc-defg-hij" (unguessable).
func NewMeetingSlug() string {
	return randString(3) + "-" + randString(4) + "-" + randString(3)
}

// NewID returns a random opaque identifier of the given length (for WS participants).
func NewID(n int) string {
	return randString(n)
}

func randString(n int) string {
	var b strings.Builder
	b.Grow(n)
	max := big.NewInt(int64(len(slugAlphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			// crypto/rand failure is fatal-ish; fall back to first char deterministically.
			b.WriteByte(slugAlphabet[0])
			continue
		}
		b.WriteByte(slugAlphabet[idx.Int64()])
	}
	return b.String()
}
