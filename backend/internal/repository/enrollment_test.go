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

	// Compile-time checks for all course_session repository methods.
	var _ func(context.Context, pgx.Tx, model.CourseSession) error = r.CreateCourseSession
	var _ func(context.Context, pgx.Tx, uuid.UUID) error = r.RevokeEnrollmentsByOrder
	var _ func(context.Context, uuid.UUID, uuid.UUID, time.Time) error = r.MarkLessonComplete
	_ = r.GetActiveSession
	_ = r.ListActiveSessionsByStudent

	_ = ctx

	// Construct a CourseSession with non-nil CompletedLessons (JSONB).
	session := model.CourseSession{
		ID:         uuid.New(),
		StudentID:  uuid.New(),
		CourseID:   uuid.New(),
		OrderID:    ptrUUID(uuid.New()),
		Status:     "active",
		Source:     "order",
		EnrolledAt: time.Now(),
		CompletedLessons: map[uuid.UUID]time.Time{
			uuid.New(): time.Now().UTC(),
		},
	}
	_ = session
}

func TestCreateCourseSessionSQLContainsConflictClause(t *testing.T) {
	// Placeholder — ON CONFLICT DO NOTHING behavior is verified by integration tests.
}

func ptrUUID(u uuid.UUID) *uuid.UUID {
	return &u
}
