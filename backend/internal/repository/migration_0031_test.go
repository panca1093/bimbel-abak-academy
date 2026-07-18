package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMigration0031_OrderParticipant(t *testing.T) {
	ctx := context.Background()
	pool := newMigration0025Pool(t)

	// Apply all migrations up to 0028 to establish the base schema.
	applyMigrationsUpTo(t, pool, "0028_multi_blank_question_audio.up.sql")

	// Pre-0031: order_participant table does not exist.
	var hasOrderParticipant bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'order_participant')`,
	).Scan(&hasOrderParticipant))
	require.False(t, hasOrderParticipant, "order_participant table must not exist before migration 0031")

	// Apply 0031 up.
	applyMigrationFile(t, pool, "0031_order_participant.up.sql")

	// FR-1: order_participant table now exists.
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'order_participant')`,
	).Scan(&hasOrderParticipant))
	require.True(t, hasOrderParticipant, "order_participant table must exist after migration 0031")

	// Setup: create a user (student role) and an order to reference.
	var studentID, orderID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO users (email, role, name) VALUES ($1, $2, $3) RETURNING id`,
		"participant-0031@test.local", "student", "Participant Test",
	).Scan(&studentID))

	// orders.student_id references users(id), so we need another user for the order's buyer.
	var buyerID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO users (email, role, name) VALUES ($1, $2, $3) RETURNING id`,
		"buyer-0031@test.local", "student", "Buyer Test",
	).Scan(&buyerID))

	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO orders (student_id) VALUES ($1) RETURNING id`, buyerID,
	).Scan(&orderID))

	// FR-2: insert a valid order_participant row.
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO order_participant (order_id, student_id) VALUES ($1, $2) RETURNING order_id`,
		orderID, studentID,
	).Scan(&orderID))

	// FR-3: duplicate (order_id, student_id) must violate PRIMARY KEY.
	_, err := pool.Exec(ctx,
		`INSERT INTO order_participant (order_id, student_id) VALUES ($1, $2)`,
		orderID, studentID,
	)
	require.Error(t, err, "duplicate (order_id, student_id) must violate PRIMARY KEY")

	// FR-4: same order with a different student is allowed.
	var student2ID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO users (email, role, name) VALUES ($1, $2, $3) RETURNING id`,
		"participant-0031-b@test.local", "student", "Participant Two",
	).Scan(&student2ID))

	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO order_participant (order_id, student_id) VALUES ($1, $2) RETURNING order_id`,
		orderID, student2ID,
	).Scan(&orderID))

	// FR-5: verify both rows exist.
	var rowCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM order_participant WHERE order_id = $1`, orderID,
	).Scan(&rowCount))
	require.Equal(t, 2, rowCount, "two participants must exist for the same order")

	// FR-6: FK violation on invalid order_id.
	_, err = pool.Exec(ctx,
		`INSERT INTO order_participant (order_id, student_id) VALUES ($1, $2)`,
		uuid.Nil, studentID,
	)
	require.Error(t, err, "FK violation on non-existent order_id must be rejected")

	// FR-7: FK violation on invalid student_id.
	_, err = pool.Exec(ctx,
		`INSERT INTO order_participant (order_id, student_id) VALUES ($1, $2)`,
		orderID, uuid.Nil,
	)
	require.Error(t, err, "FK violation on non-existent student_id must be rejected")

	// FR-8: down migration removes the table.
	applyMigrationFile(t, pool, "0031_order_participant.down.sql")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'order_participant')`,
	).Scan(&hasOrderParticipant))
	require.False(t, hasOrderParticipant, "order_participant table must be dropped by down")

	// FR-9: down migration is idempotent (running drop on a non-existent table is fine).
	applyMigrationFile(t, pool, "0031_order_participant.down.sql")
}
