package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMigration0033_ShippingOngkir(t *testing.T) {
	ctx := context.Background()
	pool := newMigration0025Pool(t)

	// Apply all migrations up to 0032 to establish the base schema.
	applyMigrationsUpTo(t, pool, "0032_unlisted_school_name.up.sql")

	// Pre-0033: shipping address columns do not exist on orders.
	var hasProvinceID, hasCityID, hasDistrictID, hasKodePos bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'province_id')`,
	).Scan(&hasProvinceID))
	require.False(t, hasProvinceID, "province_id column must not exist before migration 0033")

	// Apply 0033 up.
	applyMigrationFile(t, pool, "0033_shipping_ongkir.up.sql")

	// FR-1: shipping address columns now exist with correct types.
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'province_id' AND data_type = 'text')`,
	).Scan(&hasProvinceID))
	require.True(t, hasProvinceID, "province_id column must exist and be TEXT after migration 0033")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'city_id' AND data_type = 'text')`,
	).Scan(&hasCityID))
	require.True(t, hasCityID, "city_id column must exist and be TEXT after migration 0033")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'district_id' AND data_type = 'text')`,
	).Scan(&hasDistrictID))
	require.True(t, hasDistrictID, "district_id column must exist and be TEXT after migration 0033")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'kode_pos' AND data_type = 'text')`,
	).Scan(&hasKodePos))
	require.True(t, hasKodePos, "kode_pos column must exist and be TEXT after migration 0033")

	// FR-2: foreign key constraints exist for province_id, city_id, district_id.
	var fkCount int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM information_schema.table_constraints tc
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_name = 'orders'
		  AND tc.constraint_name LIKE '%province_id%'
	`).Scan(&fkCount))
	require.True(t, fkCount > 0, "foreign key constraint must exist for province_id")

	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM information_schema.table_constraints tc
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_name = 'orders'
		  AND tc.constraint_name LIKE '%city_id%'
	`).Scan(&fkCount))
	require.True(t, fkCount > 0, "foreign key constraint must exist for city_id")

	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM information_schema.table_constraints tc
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_name = 'orders'
		  AND tc.constraint_name LIKE '%district_id%'
	`).Scan(&fkCount))
	require.True(t, fkCount > 0, "foreign key constraint must exist for district_id")

	// FR-3: columns are nullable (can insert NULL).
	var buyerID, orderID uuid.UUID
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO users (email, role, name) VALUES ($1, $2, $3) RETURNING id`,
		"shipping-0033@test.local", "student", "Shipping Test",
	).Scan(&buyerID))

	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO orders (student_id) VALUES ($1) RETURNING id`, buyerID,
	).Scan(&orderID))

	// FR-4: can update with valid foreign keys.
	var provinceID string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT id FROM province LIMIT 1`,
	).Scan(&provinceID))

	_, updateErr := pool.Exec(ctx,
		`UPDATE orders SET province_id = $1 WHERE id = $2`,
		provinceID, orderID,
	)
	require.NoError(t, updateErr)

	// FR-5: foreign key violation is rejected.
	_, err := pool.Exec(ctx,
		`UPDATE orders SET city_id = $1 WHERE id = $2`,
		"invalid-city-id", orderID,
	)
	require.Error(t, err, "FK violation on non-existent city_id must be rejected")

	// FR-6: down migration removes all columns.
	applyMigrationFile(t, pool, "0033_shipping_ongkir.down.sql")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'province_id')`,
	).Scan(&hasProvinceID))
	require.False(t, hasProvinceID, "province_id column must be dropped by down")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'city_id')`,
	).Scan(&hasCityID))
	require.False(t, hasCityID, "city_id column must be dropped by down")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'district_id')`,
	).Scan(&hasDistrictID))
	require.False(t, hasDistrictID, "district_id column must be dropped by down")

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'kode_pos')`,
	).Scan(&hasKodePos))
	require.False(t, hasKodePos, "kode_pos column must be dropped by down")

	// FR-7: down migration is idempotent (running drop on a non-existent table is fine).
	applyMigrationFile(t, pool, "0033_shipping_ongkir.down.sql")
}
