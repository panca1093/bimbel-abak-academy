package repository

import "testing"

// Compile-time check that Repository has outbox methods.
var _ = (*Repository).InsertOutboxEvent
var _ = (*Repository).ClaimOutboxEvents
var _ = (*Repository).MarkOutboxProcessed

func TestOutboxMethodsExist(t *testing.T) {
	// Methods verified at compile time above.
}
