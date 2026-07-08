package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

// seedSectionedSession inserts a student, a utbk exam (mode='utbk'), two tests
// (with section_type set), exam_test links, a registration, and an in_progress
// exam_session. Returns the sessionID and the two testIDs (in sort order).
func seedSectionedSession(t *testing.T, env *testEnv) (sessionID string, testID1, testID2 string) {
	t.Helper()
	ctx := context.Background()

	studentID := seedUser(t, env, "student", "active", false)

	var examID string
	err := env.pool.QueryRow(ctx,
		`INSERT INTO exam (title, is_free, requires_checkin, timer_mode, result_config, mode)
		 VALUES ('UTBK Paket', true, false, 'per_test', 'hidden', 'utbk') RETURNING id`,
	).Scan(&examID)
	require.NoError(t, err)

	err = env.pool.QueryRow(ctx,
		`INSERT INTO test (title, subject, topic, duration_minutes, section_type)
		 VALUES ('TPS', 'umum', 'penalaran', 30, NULL) RETURNING id`,
	).Scan(&testID1)
	require.NoError(t, err)

	err = env.pool.QueryRow(ctx,
		`INSERT INTO test (title, subject, topic, duration_minutes, section_type)
		 VALUES ('Penalaran Matematika', 'matematika', 'kuantitatif', 45, NULL) RETURNING id`,
	).Scan(&testID2)
	require.NoError(t, err)

	_, err = env.pool.Exec(ctx,
		`INSERT INTO exam_test (exam_id, test_id, sort_order) VALUES ($1, $2, 0), ($1, $3, 1)`,
		examID, testID1, testID2,
	)
	require.NoError(t, err)

	token := "tok-" + uuid.NewString()[:8]
	var regID string
	err = env.pool.QueryRow(ctx,
		`INSERT INTO exam_registration (student_id, exam_id, token, status)
		 VALUES ($1, $2, $3, 'checked_in') RETURNING id`,
		studentID, examID, token,
	).Scan(&regID)
	require.NoError(t, err)

	err = env.pool.QueryRow(ctx,
		`INSERT INTO exam_session (registration_id, student_id, exam_id, attempt_number, started_at, status)
		 VALUES ($1, $2, $3, 1, now(), 'in_progress') RETURNING id`,
		regID, studentID, examID,
	).Scan(&sessionID)
	require.NoError(t, err)
	return sessionID, testID1, testID2
}

// TestRepo_CreateSessionSections_RoundTrip inserts two sections (first active, second pending),
// reads them back ordered by sort_order, advances the active one, and asserts the flip +
// next promotion. Then advances the last section and asserts nil next. Finally asserts that
// advancing an already-submitted section surfaces ErrNoActiveSection (the 0-row guard).
func TestRepo_CreateSessionSections_RoundTrip(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	repo := repository.New(env.pool)

	sessionIDStr, testID1Str, testID2Str := seedSectionedSession(t, env)
	sessionID := uuid.MustParse(sessionIDStr)
	testID1 := uuid.MustParse(testID1Str)
	testID2 := uuid.MustParse(testID2Str)

	// First section active with started_at=now(); second pending.
	now := time.Now().UTC()
	sections := []model.ExamSessionSection{
		{TestID: testID1, SortOrder: 0, DurationMinutes: 30, Status: "active", StartedAt: &now},
		{TestID: testID2, SortOrder: 1, DurationMinutes: 45, Status: "pending"},
	}

	tx, err := repo.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	require.NoError(t, repo.CreateSessionSectionsTx(ctx, tx, sessionID, sections))
	require.NoError(t, tx.Commit(ctx))

	// Read back ordered by sort_order.
	got, err := repo.GetSessionSections(ctx, sessionID)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, testID1, got[0].TestID)
	assert.Equal(t, 0, got[0].SortOrder)
	assert.Equal(t, "active", got[0].Status)
	assert.NotNil(t, got[0].StartedAt)
	assert.Equal(t, testID2, got[1].TestID)
	assert.Equal(t, "pending", got[1].Status)
	assert.Nil(t, got[1].StartedAt)

	// Advance the first (active) section -> flips to submitted, promotes the second.
	tx, err = repo.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	nextID, err := repo.AdvanceSessionSectionTx(ctx, tx, sessionID, testID1)
	require.NoError(t, err)
	require.NotNil(t, nextID, "next test_id must be non-nil when a pending section exists")
	assert.Equal(t, testID2, *nextID)
	require.NoError(t, tx.Commit(ctx))

	// Verify persisted state: first submitted, second active with started_at stamped.
	got, err = repo.GetSessionSections(ctx, sessionID)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "submitted", got[0].Status)
	assert.NotNil(t, got[0].SubmittedAt)
	assert.Equal(t, "active", got[1].Status)
	assert.NotNil(t, got[1].StartedAt, "promoted section must have started_at stamped")

	// Advance the last (now-active) section -> nil next (FR-12), section becomes submitted.
	tx, err = repo.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	nextID, err = repo.AdvanceSessionSectionTx(ctx, tx, sessionID, testID2)
	require.NoError(t, err)
	assert.Nil(t, nextID, "advancing the last section returns nil next test_id")
	require.NoError(t, tx.Commit(ctx))

	got, err = repo.GetSessionSections(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, "submitted", got[1].Status)

	// 0-row guard: advancing an already-submitted section surfaces ErrNoActiveSection
	// so the service (Task 3) can decide idempotent-200 vs ErrSectionNotActive.
	tx, err = repo.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = repo.AdvanceSessionSectionTx(ctx, tx, sessionID, testID1)
	assert.ErrorIs(t, err, repository.ErrNoActiveSection, "advancing a non-active section must surface the sentinel")

	// 0-row guard: advancing a test_id that was never a section also surfaces the sentinel.
	otherID := uuid.New()
	_, err = repo.AdvanceSessionSectionTx(ctx, tx, sessionID, otherID)
	assert.ErrorIs(t, err, repository.ErrNoActiveSection)
}

// TestRepo_ExtendActiveSectionTx pushes the active section's extended_until forward.
func TestRepo_ExtendActiveSectionTx(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	repo := repository.New(env.pool)

	sessionIDStr, testID1Str, _ := seedSectionedSession(t, env)
	sessionID := uuid.MustParse(sessionIDStr)
	testID1 := uuid.MustParse(testID1Str)

	now := time.Now().UTC()
	sections := []model.ExamSessionSection{
		{TestID: testID1, SortOrder: 0, DurationMinutes: 30, Status: "active", StartedAt: &now},
	}
	tx, err := repo.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	require.NoError(t, repo.CreateSessionSectionsTx(ctx, tx, sessionID, sections))
	require.NoError(t, tx.Commit(ctx))

	tx, err = repo.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	require.NoError(t, repo.ExtendActiveSectionTx(ctx, tx, sessionID, 10))
	require.NoError(t, tx.Commit(ctx))

	got, err := repo.GetSessionSections(ctx, sessionID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.NotNil(t, got[0].ExtendedUntil, "extended_until must be stamped on the active section")

	// No active section -> sentinel.
	tx, err = repo.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	_, advErr := repo.AdvanceSessionSectionTx(ctx, tx, sessionID, testID1)
	require.NoError(t, advErr) // flips active -> submitted, no next
	require.NoError(t, tx.Commit(ctx))

	tx, err = repo.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	err = repo.ExtendActiveSectionTx(ctx, tx, sessionID, 10)
	assert.ErrorIs(t, err, repository.ErrNoActiveSection, "extending with no active section surfaces the sentinel")
}