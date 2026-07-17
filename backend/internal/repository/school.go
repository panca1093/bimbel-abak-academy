package repository

import (
	"context"
	"fmt"

	"akademi-bimbel/internal/model"
)

// SchoolAdminRow is the school row returned in admin list responses,
// embedding School with a computed student_count.
type SchoolAdminRow struct {
	model.School
	StudentCount int `json:"student_count"`
}

// ListSchoolsAdmin returns all schools cursor-paginated, ordered by name.
// Cursor is the row ID (AND id > $cursor), independent of the ORDER BY column.
// Each row carries a computed student_count from the users table.
func (r *Repository) ListSchoolsAdmin(ctx context.Context, limit int, cursor string) ([]SchoolAdminRow, string, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `SELECT s.id, s.name, s.code, s.npsn, s.school_types, s.alamat,
		s.status, s.created_at, s.updated_at,
		(SELECT COUNT(*) FROM users WHERE school_id = s.id AND role = 'student' AND status != 'deleted') AS student_count
		FROM school s`
	args := []any{}
	argNum := 1

	if cursor != "" {
		query += fmt.Sprintf(` WHERE s.id > $%d::uuid`, argNum)
		args = append(args, cursor)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY s.name ASC LIMIT $%d`, argNum)
	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	schools := []SchoolAdminRow{}
	nextCursor := ""

	for rows.Next() {
		var s SchoolAdminRow
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Code, &s.NPSN, &s.SchoolTypes, &s.Alamat,
			&s.Status, &s.CreatedAt, &s.UpdatedAt, &s.StudentCount,
		); err != nil {
			return nil, "", err
		}
		if len(schools) < limit {
			schools = append(schools, s)
		} else {
			nextCursor = s.ID
		}
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	return schools, nextCursor, nil
}

// GetSchoolByID returns a school by ID. Returns nil, nil when not found.
func (r *Repository) GetSchoolByID(ctx context.Context, id string) (*model.School, error) {
	s := &model.School{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, code, npsn, school_types, alamat, status, created_at, updated_at
		FROM school WHERE id = $1`,
		id,
	).Scan(
		&s.ID, &s.Name, &s.Code, &s.NPSN, &s.SchoolTypes, &s.Alamat,
		&s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

// SchoolCodeExists checks whether a given code already exists in the school table.
// excludeID optionally excludes a specific school ID (for update checks).
func (r *Repository) SchoolCodeExists(ctx context.Context, code string, excludeID *string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM school WHERE code = $1`
	args := []any{code}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	query += `)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

// CreateSchool inserts a new school with status='active' and scans back
// id, created_at, updated_at.
func (r *Repository) CreateSchool(ctx context.Context, s *model.School) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO school (name, code, npsn, school_types, alamat, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING id, created_at, updated_at`,
		s.Name, s.Code, s.NPSN, s.SchoolTypes, s.Alamat,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

// UpdateSchool patches editable fields using COALESCE. Nil pointer arguments
// leave the corresponding column unchanged.
func (r *Repository) UpdateSchool(ctx context.Context, id string, name, npsn, alamat *string, schoolTypes []string, code *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE school
		SET name = COALESCE($1, name),
			npsn = COALESCE($2, npsn),
			alamat = COALESCE($3, alamat),
			school_types = COALESCE($4, school_types),
			code = COALESCE($5, code),
			updated_at = now()
		WHERE id = $6`,
		name, npsn, alamat, schoolTypes, code, id,
	)
	return err
}

// UpdateSchoolStatus sets the status of a school.
func (r *Repository) UpdateSchoolStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE school SET status = $1, updated_at = now() WHERE id = $2`,
		status, id,
	)
	return err
}

// GetSchoolByNameCI returns a school by its name (case-insensitive),
// or nil, nil when not found.
func (r *Repository) GetSchoolByNameCI(ctx context.Context, name string) (*model.School, error) {
	s := &model.School{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, code, npsn, school_types, alamat, status, created_at, updated_at
		FROM school WHERE LOWER(name) = LOWER($1)`,
		name,
	).Scan(
		&s.ID, &s.Name, &s.Code, &s.NPSN, &s.SchoolTypes, &s.Alamat,
		&s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

// CountStudentsBySchool returns the number of non-deleted students for a school.
func (r *Repository) CountStudentsBySchool(ctx context.Context, schoolID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE school_id = $1 AND role = 'student' AND status != 'deleted'`,
		schoolID,
	).Scan(&count)
	return count, err
}
