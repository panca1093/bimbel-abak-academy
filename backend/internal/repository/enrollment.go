package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CourseEnrollment struct {
	ID        uuid.UUID
	StudentID uuid.UUID
	ProductID uuid.UUID
	OrderID   *uuid.UUID
	Status    string
	Source    string
	EnrolledAt time.Time
	RevokedAt  *time.Time
}

type ExamRegistration struct {
	ID        uuid.UUID
	StudentID uuid.UUID
	ExamID    uuid.UUID
	OrderID   *uuid.UUID
	Token     string
	Status    string
	CreatedAt time.Time
}

func (r *Repository) CreateCourseEnrollment(ctx context.Context, tx pgx.Tx, e CourseEnrollment) error {
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

func (r *Repository) CreateExamRegistration(ctx context.Context, tx pgx.Tx, reg ExamRegistration) error {
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
