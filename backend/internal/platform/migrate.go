package platform

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"

	dbmigrations "akademi-bimbel/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, databaseURL string) error {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("open migration pool: %w", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.Glob(dbmigrations.FS, "migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(entries)

	for _, path := range entries {
		filename := path[len("migrations/"):]

		var exists bool
		err = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`,
			filename,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", filename, err)
		}
		if exists {
			continue
		}

		sql, err := fs.ReadFile(dbmigrations.FS, path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", filename, err)
		}

		if _, err = tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}

		if _, err = tx.Exec(ctx,
			`INSERT INTO schema_migrations (filename) VALUES ($1)`, filename,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", filename, err)
		}

		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", filename, err)
		}

		slog.InfoContext(ctx, "migration applied", "file", filename)
	}

	return nil
}
