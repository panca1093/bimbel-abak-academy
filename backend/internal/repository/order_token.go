package repository

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

// GenerateExamToken produces an 8-character uppercase alphanumeric token for
// an exam registration. Moved from worker/outbox.go (Task 14) so that both the
// outbox handler and the direct-grant service can call it without creating a
// circular dependency.
func GenerateExamToken() string {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read never fails on supported platforms; a constant
		// fallback would be a guessable check-in credential.
		panic(err)
	}
	return strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)[:8])
}
