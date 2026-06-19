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
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (
			email, username, phone, password_hash, role, name,
			school_id, status, otp_enabled,
			nis, dob, gender, grade, alamat_domisili, target_exam
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, $13, $14, $15
		) RETURNING id, created_at, updated_at`,
		u.Email, u.Username, u.Phone, u.PasswordHash, u.Role, u.Name,
		u.SchoolID, u.Status, u.OTPEnabled,
		u.NIS, u.DOB, u.Gender, u.Grade, u.AlamatDomisili, u.TargetExam,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	email = normalizeEmail(email)
	u := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, username, phone, password_hash, role, name,
			school_id, status, otp_enabled, created_at, updated_at,
			nis, dob, gender, grade, alamat_domisili, target_exam
		FROM users
		WHERE email = $1 AND status != 'deleted'`,
		email,
	).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.Role, &u.Name,
		&u.SchoolID, &u.Status, &u.OTPEnabled, &u.CreatedAt, &u.UpdatedAt,
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
			school_id, status, otp_enabled, created_at, updated_at,
			nis, dob, gender, grade, alamat_domisili, target_exam
		FROM users
		WHERE username = $1 AND status != 'deleted'`,
		username,
	).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.Role, &u.Name,
		&u.SchoolID, &u.Status, &u.OTPEnabled, &u.CreatedAt, &u.UpdatedAt,
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
			school_id, status, otp_enabled, created_at, updated_at,
			nis, dob, gender, grade, alamat_domisili, target_exam
		FROM users
		WHERE id = $1`,
		id,
	).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.Role, &u.Name,
		&u.SchoolID, &u.Status, &u.OTPEnabled, &u.CreatedAt, &u.UpdatedAt,
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

// UpdateUserProfile patches the editable profile fields. nil args leave the
// column unchanged via COALESCE. Email normalization is the caller's job.
func (r *Repository) UpdateUserProfile(ctx context.Context, userID string, name, email, username *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users
		SET name = COALESCE($1, name),
		    email = COALESCE($2, email),
		    username = COALESCE($3, username),
		    updated_at = now()
		WHERE id = $4`,
		name, email, username, userID,
	)
	return err
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
