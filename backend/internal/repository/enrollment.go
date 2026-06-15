package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TODO task-3: Replaced by CreateCourseSession with model.CourseSession.
// Keep as stub to unblock compile until task 3 rewrites this file.
//func (r *Repository) CreateCourseEnrollment(ctx context.Context, tx pgx.Tx, e model.CourseEnrollment) error { ... }
//func (r *Repository) RevokeEnrollmentsByOrder(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error { ... }
//func (r *Repository) CreateExamRegistration(ctx context.Context, tx pgx.Tx, reg model.ExamRegistration) error { ... }
//func (r *Repository) ExpireExamRegistrationsByOrder(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error { ... }

func (r *Repository) RevokeEnrollmentsByOrder(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE course_enrollment SET status = 'revoked', revoked_at = now()
		WHERE order_id = $1`,
		orderID,
	)
	return err
}
