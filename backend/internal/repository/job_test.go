package repository

import (
	"context"

	"akademi-bimbel/internal/model"
)

// Compile-time check: *Repository must implement all job methods.
// Behavioral coverage lands in Task 9's service-layer tests (testcontainers
// Postgres), matching this package's existing convention (see order_test.go).
var _ interface {
	CreateJob(context.Context, *model.Job) error
	GetJobByID(context.Context, string) (*model.Job, error)
	ClaimNextJob(context.Context) (*model.Job, error)
	UpdateJobProgress(context.Context, string, int) error
	FinishJob(context.Context, string, string, int, *string, *string) error
} = (*Repository)(nil)
