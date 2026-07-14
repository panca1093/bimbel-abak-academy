package repository

import (
	"context"
	"fmt"
	"time"

	"akademi-bimbel/internal/model"
)

// StudentRow is the student shape returned in admin school student list
// responses (no password_hash, no student-only fields beyond nis/grade).
type StudentRow struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	NIS       string    `json:"nis"`
	Email     *string   `json:"email"`
	Status    string    `json:"status"`
	Grade     *int      `json:"grade"`
	CreatedAt time.Time `json:"created_at"`
}

// StudentFilter carries optional filters for ListStudentsBySchool.
type StudentFilter struct {
	Status string
	Cursor string
	Limit  int
	Q      string
}

// ListStudentsBySchool returns non-deleted students scoped to a school,
// cursor-paginated (same shape as ListAdminUsers). Supports optional
// status filter and free-text search on name/nis.
func (r *Repository) ListStudentsBySchool(ctx context.Context, schoolID string, filter StudentFilter) ([]StudentRow, string, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	query := `SELECT id, name, username, nis, email, status, grade, created_at
		FROM users WHERE school_id = $1 AND role = 'student' AND status != 'deleted'`
	args := []any{schoolID}
	argNum := 2

	if filter.Status != "" {
		query += fmt.Sprintf(` AND status = $%d`, argNum)
		args = append(args, filter.Status)
		argNum++
	}
	if filter.Q != "" {
		query += fmt.Sprintf(` AND (name ILIKE $%d OR nis ILIKE $%d)`, argNum, argNum)
		args = append(args, "%"+filter.Q+"%")
		argNum++
	}
	if filter.Cursor != "" {
		query += fmt.Sprintf(` AND id < $%d::uuid`, argNum)
		args = append(args, filter.Cursor)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, argNum)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	students := []StudentRow{}
	nextCursor := ""

	for rows.Next() {
		var s StudentRow
		if err := rows.Scan(&s.ID, &s.Name, &s.Username, &s.NIS, &s.Email, &s.Status, &s.Grade, &s.CreatedAt); err != nil {
			return nil, "", err
		}
		if len(students) < filter.Limit {
			students = append(students, s)
		} else {
			nextCursor = s.ID
		}
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	return students, nextCursor, nil
}

// CreateStudent inserts a new user with role='student', otp_enabled=false,
// and scans back id, created_at, updated_at.
func (r *Repository) CreateStudent(ctx context.Context, u *model.User) error {
	if u.Email != nil {
		normalized := normalizeEmail(*u.Email)
		u.Email = &normalized
	}
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (
			email, username, password_hash, role, name,
			school_id, status, otp_enabled,
			nis, dob, gender, grade, alamat_domisili, target_exam
		) VALUES (
			$1, $2, $3, 'student', $4,
			$5, 'active', false,
			$6, $7, $8, $9, $10, $11
		) RETURNING id, created_at, updated_at`,
		u.Email, u.Username, u.PasswordHash, u.Name,
		u.SchoolID,
		u.NIS, u.DOB, u.Gender, u.Grade, u.AlamatDomisili, u.TargetExam,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// GetStudentByID returns a student by ID scoped to a specific school.
// Returns nil, nil when not found (including when the student belongs to
// a different school — indistinguishable from "doesn't exist").
func (r *Repository) GetStudentByID(ctx context.Context, id, schoolID string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, username, phone, password_hash, role, name,
			school_id, photo_url, status, otp_enabled, auth_provider, created_at, updated_at,
			nis, dob, gender, grade, alamat_domisili, target_exam
		FROM users
		WHERE id = $1 AND school_id = $2 AND role = 'student' AND status != 'deleted'`,
		id, schoolID,
	).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.Role, &u.Name,
		&u.SchoolID, &u.PhotoURL, &u.Status, &u.OTPEnabled, &u.AuthProvider, &u.CreatedAt, &u.UpdatedAt,
		&u.NIS, &u.DOB, &u.Gender, &u.Grade, &u.AlamatDomisili, &u.TargetExam,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

// UpdateStudentStatus sets the status of a student, scoped to a specific school.
// Returns the number of rows affected (0 if the student doesn't exist or
// doesn't belong to the given school).
func (r *Repository) UpdateStudentStatus(ctx context.Context, id, schoolID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET status = $1, updated_at = now()
		WHERE id = $2 AND school_id = $3 AND role = 'student' AND status != 'deleted'`,
		status, id, schoolID,
	)
	return err
}

// ResetStudentPasswordHash overwrites the password hash for a student,
// scoped to a specific school.
func (r *Repository) ResetStudentPasswordHash(ctx context.Context, id, schoolID, hash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now()
		WHERE id = $2 AND school_id = $3 AND role = 'student' AND status != 'deleted'`,
		hash, id, schoolID,
	)
	return err
}
