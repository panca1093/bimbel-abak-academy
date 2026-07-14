package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMigration0026_ExamProductDecouple(t *testing.T) {
	ctx := context.Background()
	pool := newMigration0025Pool(t)

	// Bring the DB to exactly the 0025 schema.
	applyMigrationsUpTo(t, pool, "0025_question_bank.up.sql")

	var linkedProductID uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO product (type, name, price, status) VALUES ('exam', $1, 0, 'draft') RETURNING id`,
		"Linked Exam Product",
	).Scan(&linkedProductID)
	require.NoError(t, err)

	var linkedExamID, freeExamID uuid.UUID
	err = pool.QueryRow(ctx,
		`INSERT INTO exam (title, status, product_id) VALUES ($1, 'draft', $2) RETURNING id`,
		"Linked Exam", linkedProductID,
	).Scan(&linkedExamID)
	require.NoError(t, err)

	err = pool.QueryRow(ctx,
		`INSERT INTO exam (title, status) VALUES ($1, 'draft') RETURNING id`,
		"Free Exam",
	).Scan(&freeExamID)
	require.NoError(t, err)

	// Apply 0026 up.
	applyMigrationFile(t, pool, "0026_exam_product_decouple.up.sql")

	// The existing 1:1 link must survive as a product_exam row.
	var joinCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_exam WHERE product_id = $1 AND exam_id = $2`,
		linkedProductID, linkedExamID,
	).Scan(&joinCount))
	require.Equal(t, 1, joinCount, "pre-existing exam.product_id link must be preserved in product_exam")

	// The free exam has no product link and must not gain a spurious join row.
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_exam WHERE exam_id = $1`, freeExamID,
	).Scan(&joinCount))
	require.Equal(t, 0, joinCount)

	// exam.product_id column must be gone.
	var hasProductID bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'exam' AND column_name = 'product_id')`,
	).Scan(&hasProductID))
	require.False(t, hasProductID, "exam.product_id must be dropped by 0026 up")

	// M:N now holds: attach a second exam to the same product.
	var secondExamID uuid.UUID
	err = pool.QueryRow(ctx,
		`INSERT INTO exam (title, status) VALUES ($1, 'draft') RETURNING id`,
		"Second Exam On Same Product",
	).Scan(&secondExamID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx,
		`INSERT INTO product_exam (product_id, exam_id) VALUES ($1, $2)`,
		linkedProductID, secondExamID,
	)
	require.NoError(t, err, "a product must be attachable to more than one exam")

	// Determine the expected deterministic winner (lowest exam_id) before down
	// drops product_exam entirely.
	var winnerID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT exam_id FROM product_exam WHERE product_id = $1 ORDER BY exam_id ASC LIMIT 1`,
		linkedProductID,
	).Scan(&winnerID))

	// Apply 0026 down.
	applyMigrationFile(t, pool, "0026_exam_product_decouple.down.sql")

	var hasProductIDAfterDown bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'exam' AND column_name = 'product_id')`,
	).Scan(&hasProductIDAfterDown))
	require.True(t, hasProductIDAfterDown, "exam.product_id must be restored by down")

	var winnerProductID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT product_id FROM exam WHERE id = $1`, winnerID,
	).Scan(&winnerProductID))
	require.Equal(t, linkedProductID, winnerProductID)

	loserID := linkedExamID
	if winnerID == linkedExamID {
		loserID = secondExamID
	}
	var loserProductID *uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT product_id FROM exam WHERE id = $1`, loserID,
	).Scan(&loserProductID))
	require.Nil(t, loserProductID, "the exam not chosen by the deterministic collapse must have a NULL product_id")

	// Free exam must still have no product_id.
	var freeExamProductID *uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT product_id FROM exam WHERE id = $1`, freeExamID,
	).Scan(&freeExamProductID))
	require.Nil(t, freeExamProductID)

	// uq_exam_product must be restored, and must tolerate multiple NULL
	// product_id rows (Postgres treats NULLs as distinct in a partial unique
	// index) — proven by the fact that both freeExamID and loserID already
	// have NULL product_id above without conflict.
	var uniqueIndexExists bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE indexname = 'uq_exam_product')`,
	).Scan(&uniqueIndexExists))
	require.True(t, uniqueIndexExists, "uq_exam_product must be restored by down")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'product_exam')`,
	).Scan(&hasProductID))
	require.False(t, hasProductID, "product_exam must be dropped by down")
}
