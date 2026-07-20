package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestAllocateCertificateNumber proves FR-9/FR-10/FR-11: a number is minted on first
// call, repeated calls on the same session return the identical number without
// consuming another sequence value, and distinct sessions get distinct numbers.
func TestAllocateCertificateNumber(t *testing.T) {
	pool := newGradingTestPool(t)
	repo := New(pool)
	ctx := context.Background()

	studentA := insertGradingUser(t, pool, "student", "Student CertNum A")
	studentB := insertGradingUser(t, pool, "student", "Student CertNum B")
	testID := insertGradingTest(t, pool)
	examID := insertGradingExam(t, pool, testID)

	sessionA := insertGradingSession(t, pool, studentA, examID, "submitted", nil, nil)
	sessionB := insertGradingSession(t, pool, studentB, examID, "submitted", nil, nil)

	seqValue := func() int64 {
		t.Helper()
		var v int64
		require.NoError(t, pool.QueryRow(ctx, `SELECT last_value FROM certificate_number_seq`).Scan(&v))
		return v
	}

	t.Run("first call allocates a number in the ABK/YYYY/NNNNNN shape", func(t *testing.T) {
		number, err := repo.AllocateCertificateNumber(ctx, sessionA)
		if err != nil {
			t.Fatalf("AllocateCertificateNumber: %v", err)
		}
		if !strings.HasPrefix(number, "ABK/") {
			t.Errorf("number %q should start with ABK/", number)
		}
		parts := strings.Split(number, "/")
		if len(parts) != 3 {
			t.Fatalf("number %q should have 3 slash-separated parts, got %d", number, len(parts))
		}
		if len(parts[1]) != 4 {
			t.Errorf("year part %q should be 4 digits", parts[1])
		}
		if len(parts[2]) != 6 {
			t.Errorf("sequence part %q should be zero-padded to 6 digits", parts[2])
		}

		// Persisted on the session row.
		sess, err := repo.GetExamSessionByID(ctx, sessionA)
		if err != nil {
			t.Fatalf("GetExamSessionByID: %v", err)
		}
		if sess.CertificateNumber == nil || *sess.CertificateNumber != number {
			t.Errorf("CertificateNumber = %v, want %q", sess.CertificateNumber, number)
		}
	})

	t.Run("repeated calls are idempotent and do not consume the sequence again", func(t *testing.T) {
		first, err := repo.AllocateCertificateNumber(ctx, sessionA)
		if err != nil {
			t.Fatalf("AllocateCertificateNumber (1st in subtest): %v", err)
		}
		seqBefore := seqValue()

		second, err := repo.AllocateCertificateNumber(ctx, sessionA)
		if err != nil {
			t.Fatalf("AllocateCertificateNumber (2nd): %v", err)
		}
		if second != first {
			t.Errorf("second call returned a different number: %q vs %q", second, first)
		}

		third, err := repo.AllocateCertificateNumber(ctx, sessionA)
		if err != nil {
			t.Fatalf("AllocateCertificateNumber (3rd): %v", err)
		}
		if third != first {
			t.Errorf("third call returned a different number: %q vs %q", third, first)
		}

		seqAfter := seqValue()
		if seqAfter != seqBefore {
			t.Errorf("sequence advanced on repeat calls: before=%d after=%d", seqBefore, seqAfter)
		}
	})

	t.Run("distinct sessions get distinct numbers", func(t *testing.T) {
		numberA, err := repo.AllocateCertificateNumber(ctx, sessionA)
		if err != nil {
			t.Fatalf("AllocateCertificateNumber sessionA: %v", err)
		}
		numberB, err := repo.AllocateCertificateNumber(ctx, sessionB)
		if err != nil {
			t.Fatalf("AllocateCertificateNumber sessionB: %v", err)
		}
		if numberA == numberB {
			t.Errorf("sessionA and sessionB got the same number %q", numberA)
		}
	})

	t.Run("unknown session returns ErrNotFound", func(t *testing.T) {
		_, err := repo.AllocateCertificateNumber(ctx, uuid.New())
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("want ErrNotFound, got %v", err)
		}
	})
}

// Compile-time check: *Repository implements AllocateCertificateNumber.
var _ interface {
	AllocateCertificateNumber(context.Context, uuid.UUID) (string, error)
} = (*Repository)(nil)
