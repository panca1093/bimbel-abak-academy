package infra

import "testing"

// Compile-time signature checks.
var _ = SaveIdempotentResponse
var _ = GetIdempotentResponse

func TestIdempotencyFunctionsExist(t *testing.T) {
	// Verified at compile time.
}
