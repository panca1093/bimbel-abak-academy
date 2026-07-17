package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func columnExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table, column string) bool {
	t.Helper()
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = $1 AND column_name = $2)`,
		table, column,
	).Scan(&exists)
	require.NoError(t, err)
	return exists
}

func columnNullable(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table, column string) bool {
	t.Helper()
	var nullable string
	err := pool.QueryRow(ctx,
		`SELECT is_nullable FROM information_schema.columns WHERE table_name = $1 AND column_name = $2`,
		table, column,
	).Scan(&nullable)
	require.NoError(t, err)
	return nullable == "YES"
}

func TestMigration0030_UserBiodataChanges(t *testing.T) {
	pool := newMigration0025Pool(t)

	// Apply all migrations up to and including 0029 (province/city/district tables).
	applyMigrationsUpTo(t, pool, "0029_seed_province_city_district.up.sql")

	// ---- Apply 0030 UP ----
	applyMigrationFile(t, pool, "0030_user_biodata_changes.up.sql")

	ctx := context.Background()

	// Verify: nis is dropped.
	require.False(t, columnExists(t, ctx, pool, "users", "nis"),
		"nis column must be dropped by up migration")

	// Verify: five new columns exist.
	require.True(t, columnExists(t, ctx, pool, "users", "jenjang"),
		"jenjang column must exist after up migration")
	require.True(t, columnExists(t, ctx, pool, "users", "provinsi_id"),
		"provinsi_id column must exist after up migration")
	require.True(t, columnExists(t, ctx, pool, "users", "kota_id"),
		"kota_id column must exist after up migration")
	require.True(t, columnExists(t, ctx, pool, "users", "kecamatan_id"),
		"kecamatan_id column must exist after up migration")
	require.True(t, columnExists(t, ctx, pool, "users", "kode_pos"),
		"kode_pos column must exist after up migration")

	// Verify: all new columns are nullable.
	for _, col := range []string{"jenjang", "provinsi_id", "kota_id", "kecamatan_id", "kode_pos"} {
		require.True(t, columnNullable(t, ctx, pool, "users", col),
			"%s must be nullable", col)
	}

	// Verify FK constraints exist for the three reference columns.
	var constraintCount int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_name = 'users'
		  AND tc.constraint_type = 'FOREIGN KEY'
		  AND kcu.column_name IN ('provinsi_id', 'kota_id', 'kecamatan_id')
	`).Scan(&constraintCount)
	require.NoError(t, err)
	require.Equal(t, 3, constraintCount,
		"provinsi_id, kota_id, kecamatan_id must each have an FK constraint")

	// ---- Apply 0030 DOWN ----
	applyMigrationFile(t, pool, "0030_user_biodata_changes.down.sql")

	// Verify: nis is restored and nullable.
	require.True(t, columnExists(t, ctx, pool, "users", "nis"),
		"nis column must be restored by down migration")
	require.True(t, columnNullable(t, ctx, pool, "users", "nis"),
		"nis must remain nullable after down migration")

	// Verify: the five new columns are dropped.
	require.False(t, columnExists(t, ctx, pool, "users", "jenjang"),
		"jenjang column must be dropped by down migration")
	require.False(t, columnExists(t, ctx, pool, "users", "provinsi_id"),
		"provinsi_id column must be dropped by down migration")
	require.False(t, columnExists(t, ctx, pool, "users", "kota_id"),
		"kota_id column must be dropped by down migration")
	require.False(t, columnExists(t, ctx, pool, "users", "kecamatan_id"),
		"kecamatan_id column must be dropped by down migration")
	require.False(t, columnExists(t, ctx, pool, "users", "kode_pos"),
		"kode_pos column must be dropped by down migration")
}
