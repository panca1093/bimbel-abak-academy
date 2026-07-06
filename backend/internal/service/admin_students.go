package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateNIS      = errors.New("nis already registered in this school")
	ErrSchoolDeactivated  = errors.New("school is deactivated")
	ErrStudentNotFound    = errors.New("student not found")
)

const tempPasswordLen = 10

var tempPasswordChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func genTempPassword() (string, error) {
	pass := make([]rune, tempPasswordLen)
	for i := range pass {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(tempPasswordChars))))
		if err != nil {
			return "", err
		}
		pass[i] = tempPasswordChars[n.Int64()]
	}
	return string(pass), nil
}

// --- response types ---

type StudentRegistrationResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Username     string  `json:"username"`
	NIS          string  `json:"nis"`
	Email        *string `json:"email"`
	TempPassword string  `json:"temp_password"`
	CreatedAt    string  `json:"created_at"`
}

type StudentResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Username  string  `json:"username"`
	NIS       string  `json:"nis"`
	Email     *string `json:"email"`
	Status    string  `json:"status"`
	Grade     *int    `json:"grade"`
	CreatedAt string  `json:"created_at"`
}

type StudentCredentialsResponse struct {
	Username     string `json:"username"`
	TempPassword string `json:"temp_password"`
}

func toStudentResponse(row repository.StudentRow) StudentResponse {
	var grade *int
	if row.Grade != nil {
		grade = row.Grade
	}
	return StudentResponse{
		ID:        row.ID,
		Name:      row.Name,
		Username:  row.Username,
		NIS:       row.NIS,
		Email:     row.Email,
		Status:    row.Status,
		Grade:     grade,
		CreatedAt: row.CreatedAt.Format(time.RFC3339),
	}
}

// --- methods ---

// RegisterStudent creates a new student user under the given school.
// Returns the plaintext temp password exactly once in the response.
func (s *Service) RegisterStudent(ctx context.Context, schoolID, name, nis string, email *string, dob *time.Time, gender *string, grade *int, alamatDomisili, targetExam *string) (*StudentRegistrationResponse, error) {
	if name == "" || nis == "" {
		return nil, ErrMissingField
	}

	school, err := s.storeRepo.GetSchoolByID(ctx, schoolID)
	if err != nil {
		return nil, err
	}
	if school == nil {
		return nil, ErrSchoolNotFound
	}
	if school.Status == "deactivated" {
		return nil, ErrSchoolDeactivated
	}

	username := school.Code + "_" + nis

	existing, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateNIS
	}

	tempPass, err := genTempPassword()
	if err != nil {
		return nil, fmt.Errorf("generate temp password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(tempPass), 12)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:       &username,
		Name:           name,
		Email:          email,
		PasswordHash:   string(hash),
		Role:           RoleStudent,
		SchoolID:       &schoolID,
		Status:         "active",
		OTPEnabled:     false,
		NIS:            &nis,
		DOB:            dob,
		Gender:         gender,
		Grade:          grade,
		AlamatDomisili: alamatDomisili,
		TargetExam:     targetExam,
	}
	if err := s.storeRepo.CreateStudent(ctx, user); err != nil {
		return nil, err
	}

	return &StudentRegistrationResponse{
		ID:           user.ID,
		Name:         user.Name,
		Username:     username,
		NIS:          nis,
		Email:        email,
		TempPassword: tempPass,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ListStudents returns cursor-paginated students scoped to the given school.
func (s *Service) ListStudents(ctx context.Context, schoolID string, statusFilter, q string, limit int, cursor string) ([]StudentResponse, string, error) {
	rows, nextCursor, err := s.storeRepo.ListStudentsBySchool(ctx, schoolID, repository.StudentFilter{
		Status: statusFilter,
		Cursor: cursor,
		Limit:  limit,
		Q:      q,
	})
	if err != nil {
		return nil, "", err
	}

	students := make([]StudentResponse, len(rows))
	for i, r := range rows {
		students[i] = toStudentResponse(r)
	}
	return students, nextCursor, nil
}

// ChangeStudentStatus toggles a student's active/deactivated status.
// Row-scoping via schoolID + student ID — returns ErrStudentNotFound if
// the student does not exist or belongs to a different school.
func (s *Service) ChangeStudentStatus(ctx context.Context, schoolID, targetID, newStatus string) error {
	if newStatus != "active" && newStatus != "deactivated" {
		return fmt.Errorf("%w: %s", ErrInvalidStatusFilter, newStatus)
	}

	student, err := s.storeRepo.GetStudentByID(ctx, targetID, schoolID)
	if err != nil {
		return err
	}
	if student == nil {
		return ErrStudentNotFound
	}
	return s.storeRepo.UpdateStudentStatus(ctx, targetID, schoolID, newStatus)
}

// ReissueStudentCredentials generates a new temp password, overwrites the
// stored hash, and returns the plaintext password exactly once.
func (s *Service) ReissueStudentCredentials(ctx context.Context, schoolID, targetID string) (*StudentCredentialsResponse, error) {
	student, err := s.storeRepo.GetStudentByID(ctx, targetID, schoolID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, ErrStudentNotFound
	}

	tempPass, err := genTempPassword()
	if err != nil {
		return nil, fmt.Errorf("generate temp password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(tempPass), 12)
	if err != nil {
		return nil, err
	}

	if err := s.storeRepo.ResetStudentPasswordHash(ctx, targetID, schoolID, string(hash)); err != nil {
		return nil, err
	}

	return &StudentCredentialsResponse{
		Username:     *student.Username,
		TempPassword: tempPass,
	}, nil
}
