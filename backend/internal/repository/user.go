package repository

import (
	"context"
	"strings"

	"akademi-bimbel/internal/model"
)

func normalizeEmail(email string) string {
	return strings.ToLower(email)
}

func (r *Repository) CreateUser(ctx context.Context, u *model.User) error {
	if u.Email != nil {
		normalized := normalizeEmail(*u.Email)
		u.Email = &normalized
	}
	if u.AuthProvider == "" {
		u.AuthProvider = "password"
	}
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (
			email, username, phone, password_hash, role, name,
			school_id, photo_url, status, otp_enabled, auth_provider,
			nis, dob, gender, grade, alamat_domisili, target_exam
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17
		) RETURNING id, created_at, updated_at`,
		u.Email, u.Username, u.Phone, u.PasswordHash, u.Role, u.Name,
		u.SchoolID, u.PhotoURL, u.Status, u.OTPEnabled, u.AuthProvider,
		u.NIS, u.DOB, u.Gender, u.Grade, u.AlamatDomisili, u.TargetExam,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	email = normalizeEmail(email)
	u := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, username, phone, password_hash, role, name,
			school_id, photo_url, status, otp_enabled, auth_provider, created_at, updated_at,
			nis, dob, gender, grade, alamat_domisili, target_exam
		FROM users
		WHERE email = $1 AND status != 'deleted'`,
		email,
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

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, username, phone, password_hash, role, name,
			school_id, photo_url, status, otp_enabled, auth_provider, created_at, updated_at,
			nis, dob, gender, grade, alamat_domisili, target_exam
		FROM users
		WHERE username = $1 AND status != 'deleted'`,
		username,
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

func (r *Repository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, username, phone, password_hash, role, name,
			school_id, photo_url, status, otp_enabled, auth_provider, created_at, updated_at,
			nis, dob, gender, grade, alamat_domisili, target_exam
		FROM users
		WHERE id = $1`,
		id,
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

func (r *Repository) UpdatePasswordHash(ctx context.Context, userID, hash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`,
		hash, userID,
	)
	return err
}

// ActivateUser transitions a pending_verification user to active in a single
// UPDATE, so verification is atomic (no read-modify-write).
func (r *Repository) ActivateUser(ctx context.Context, userID string) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET status = 'active', otp_enabled = false, updated_at = now()
		WHERE id = $1 AND status = 'pending_verification'`,
		userID,
	)
	return tag.RowsAffected() == 1, err
}

// UpdateUserProfile patches the editable profile fields. nil args leave the
// column unchanged via COALESCE. Email normalization is the caller's job.
func (r *Repository) UpdateUserProfile(ctx context.Context, userID string, name, email, username, phone, address, targetExam *string, grade *int, schoolID *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users
		SET name = COALESCE($1, name),
		    email = COALESCE($2, email),
		    username = COALESCE($3, username),
		    phone = COALESCE($4, phone),
		    alamat_domisili = COALESCE($5, alamat_domisili),
		    target_exam = COALESCE($6, target_exam),
		    grade = COALESCE($7, grade),
		    school_id = COALESCE($8, school_id),
		    updated_at = now()
		WHERE id = $9`,
		name, email, username, phone, address, targetExam, grade, schoolID, userID,
	)
	return err
}

// UpdateUserPhoto sets the user's avatar URL.
func (r *Repository) UpdateUserPhoto(ctx context.Context, userID, photoURL string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET photo_url = $1, updated_at = now() WHERE id = $2`,
		photoURL, userID,
	)
	return err
}

// ListSchools returns active schools ordered by name.
func (r *Repository) ListSchools(ctx context.Context) ([]*model.School, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, code, school_types FROM school WHERE status = 'active' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schools := []*model.School{}
	for rows.Next() {
		s := &model.School{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.SchoolTypes); err != nil {
			return nil, err
		}
		schools = append(schools, s)
	}
	return schools, rows.Err()
}

func (r *Repository) TombstoneUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users
		SET status = 'deleted', name = '[deleted user]',
		    email = NULL, phone = NULL, alamat_domisili = NULL,
		    updated_at = now()
		WHERE id = $1`,
		userID,
	)
	return err
}
