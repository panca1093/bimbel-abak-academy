package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
)

func TestEnrollmentMethods(t *testing.T) {
	r := &Repository{}
	ctx := context.Background()

	_ = r.RevokeEnrollmentsByOrder

	// TODO task-3: Add compile-time checks for CreateCourseSession.
	// The old CreateCourseEnrollment, CreateExamRegistration, ExpireExamRegistrationsByOrder
	// were removed with their model types; task 3 rewrites this file.

	enrollment := model.CourseSession{
		ID:        uuid.New(),
		StudentID: uuid.New(),
		CourseID:  uuid.New(),
		OrderID:   ptrUUID(uuid.New()),
		Status:    "active",
		Source:    "order",
		EnrolledAt: time.Now(),
	}
	_ = enrollment

	_ = ctx

	// Compile-time check: RevokeEnrollmentsByOrder still active
	var _ func(context.Context, pgx.Tx, uuid.UUID) error = r.RevokeEnrollmentsByOrder
}

func TestCreateCourseEnrollmentSQLContainsConflictClause(t *testing.T) {
	// TODO task-3: Rewrite for course_session.
	// Verification that CreateCourseSession uses ON CONFLICT DO NOTHING.
	// Currently stubbed until task 3.
}

func TestCreateExamRegistrationSQLContainsConflictClause(t *testing.T) {
	// Deferred to phase 3 — method removed.
}

func ptrUUID(u uuid.UUID) *uuid.UUID {
	return &u
}
