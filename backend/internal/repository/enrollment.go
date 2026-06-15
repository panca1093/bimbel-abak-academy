package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
)

// CreateCourseSession inserts a new course_session with ON CONFLICT DO NOTHING.
// Idempotent for (student_id, course_id) WHERE status='active'.
func (r *Repository) CreateCourseSession(ctx context.Context, tx pgx.Tx, s model.CourseSession) error {
	lessonsJSON, err := serializeCompletedLessons(s.CompletedLessons)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO course_session (student_id, course_id, order_id, status, source, enrolled_at, completed_lessons)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING`,
		s.StudentID, s.CourseID, s.OrderID, s.Status, s.Source, s.EnrolledAt, lessonsJSON,
	)
	return err
}

// RevokeEnrollmentsByOrder revokes all course_sessions for a given order.
func (r *Repository) RevokeEnrollmentsByOrder(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE course_session SET status = 'revoked', revoked_at = now()
		WHERE order_id = $1`,
		orderID,
	)
	return err
}

// MarkLessonComplete marks a lesson as complete in the course_session.
// First timestamp wins — re-marking the same lesson is a no-op, not an error.
func (r *Repository) MarkLessonComplete(ctx context.Context, sessionID uuid.UUID, lessonID uuid.UUID, at time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE course_session
		SET completed_lessons = completed_lessons || jsonb_build_object($2, $3)
		WHERE id = $1 AND NOT (completed_lessons ? $2)`,
		sessionID, lessonID.String(), at.Format(time.RFC3339Nano),
	)
	return err
}

// GetActiveSession returns the active course_session for a student + course pair.
// Returns ErrNotFound when no active session exists.
func (r *Repository) GetActiveSession(ctx context.Context, studentID uuid.UUID, courseID uuid.UUID) (model.CourseSession, error) {
	var s model.CourseSession
	var lessonBytes []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, course_id, order_id, status, source, enrolled_at, revoked_at, completed_lessons
		FROM course_session
		WHERE student_id = $1 AND course_id = $2 AND status = 'active'`,
		studentID, courseID,
	).Scan(&s.ID, &s.StudentID, &s.CourseID, &s.OrderID, &s.Status, &s.Source, &s.EnrolledAt, &s.RevokedAt, &lessonBytes)
	if err != nil {
		if isNotFound(err) {
			return model.CourseSession{}, ErrNotFound
		}
		return model.CourseSession{}, err
	}
	if err := deserializeCompletedLessons(lessonBytes, &s.CompletedLessons); err != nil {
		return model.CourseSession{}, err
	}
	return s, nil
}

// ListActiveSessionsByStudent returns all active course_sessions for a student.
func (r *Repository) ListActiveSessionsByStudent(ctx context.Context, studentID uuid.UUID) ([]model.CourseSession, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, student_id, course_id, order_id, status, source, enrolled_at, revoked_at, completed_lessons
		FROM course_session
		WHERE student_id = $1 AND status = 'active'`,
		studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.CourseSession
	for rows.Next() {
		var s model.CourseSession
		var lessonBytes []byte
		if err := rows.Scan(&s.ID, &s.StudentID, &s.CourseID, &s.OrderID, &s.Status, &s.Source, &s.EnrolledAt, &s.RevokedAt, &lessonBytes); err != nil {
			return nil, err
		}
		if err := deserializeCompletedLessons(lessonBytes, &s.CompletedLessons); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

// serializeCompletedLessons marshals the completed_lessons map to JSON bytes.
func serializeCompletedLessons(m map[uuid.UUID]time.Time) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// deserializeCompletedLessons unmarshals JSONB bytes into the completed_lessons map.
func deserializeCompletedLessons(data []byte, m *map[uuid.UUID]time.Time) error {
	if len(data) == 0 || string(data) == "null" {
		*m = make(map[uuid.UUID]time.Time)
		return nil
	}
	return json.Unmarshal(data, m)
}
