package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMigration0027_MerchandiseProductType(t *testing.T) {
	ctx := context.Background()
	pool := newMigration0025Pool(t)

	applyMigrationsUpTo(t, pool, "0026_exam_product_decouple.up.sql")

	// FR-1: pre-0027, the narrow CHECK rejects 'merchandise'.
	_, err := pool.Exec(ctx,
		`INSERT INTO product (type, name, price) VALUES ('merchandise', 'Pre-migration Tee', 100)`,
	)
	require.Error(t, err, "merchandise must be rejected before the widening migration")

	applyMigrationFile(t, pool, "0027_merchandise_product_type.up.sql")

	// FR-1: after up, inserting a merchandise product succeeds.
	var merchID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO product (type, name, price) VALUES ('merchandise', 'Academy Tee', 100) RETURNING id`,
	).Scan(&merchID))
	var medalID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO product (type, name, price) VALUES ('medal', 'Gold Medal', 100) RETURNING id`,
	).Scan(&medalID))

	// Control: an unknown type still violates the widened CHECK.
	_, err = pool.Exec(ctx,
		`INSERT INTO product (type, name, price) VALUES ('bogus', 'Bogus', 100)`,
	)
	require.Error(t, err, "an unknown type must still violate the widened CHECK")

	// FR-2: down with a pre-existing merchandise row is fail-safe — runs without
	// error and converts the row to 'book' rather than orphaning it.
	applyMigrationFile(t, pool, "0027_merchandise_product_type.down.sql")

	var typeAfterDown string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT type FROM product WHERE id = $1`, merchID,
	).Scan(&typeAfterDown))
	require.Equal(t, "book", typeAfterDown, "down must convert merchandise rows to book, not drop them")
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT type FROM product WHERE id = $1`, medalID,
	).Scan(&typeAfterDown))
	require.Equal(t, "book", typeAfterDown, "down must convert medal rows to book, not drop them")

	// After down, the narrow CHECK is back: merchandise is rejected again.
	_, err = pool.Exec(ctx,
		`INSERT INTO product (type, name, price) VALUES ('merchandise', 'Post-down Tee', 100)`,
	)
	require.Error(t, err, "narrow CHECK must be restored after down")
	_, err = pool.Exec(ctx,
		`INSERT INTO product (type, name, price) VALUES ('medal', 'Post-down Medal', 100)`,
	)
	require.Error(t, err, "narrow CHECK must reject medals after down")
}
