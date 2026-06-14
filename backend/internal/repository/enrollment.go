package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
)

func (r *Repository) CreateCourseEnrollment(ctx context.Context, tx pgx.Tx, e model.CourseEnrollment) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO course_enrollment (student_id, product_id, order_id, status, source, enrolled_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		e.StudentID, e.ProductID, e.OrderID, e.Status, e.Source, e.EnrolledAt,
	)
	return err
}

func (r *Repository) RevokeEnrollmentsByOrder(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE course_enrollment SET status = 'revoked', revoked_at = now()
		WHERE order_id = $1`,
		orderID,
	)
	return err
}

func (r *Repository) CreateExamRegistration(ctx context.Context, tx pgx.Tx, reg model.ExamRegistration) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO exam_registration (student_id, exam_id, order_id, token, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING`,
		reg.StudentID, reg.ExamID, reg.OrderID, reg.Token, reg.Status,
	)
	return err
}

func (r *Repository) ExpireExamRegistrationsByOrder(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE exam_registration SET status = 'expired'
		WHERE order_id = $1`,
		orderID,
	)
	return err
}
