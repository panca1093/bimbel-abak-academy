package repository

import (
	"context"
)

// Compile-time check: *Repository must implement all system config methods.
var _ interface {
	ListSystemConfig(context.Context) ([]SystemConfigRow, error)
	UpsertSystemConfig(context.Context, string, string, bool) error
} = (*Repository)(nil)
