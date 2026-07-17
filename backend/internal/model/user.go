package model

import "time"

type User struct {
	ID           string     `json:"id"`
	Email        *string    `json:"email"`
	Username     *string    `json:"username"`
	Phone        *string    `json:"phone"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	Name         string     `json:"name"`
	SchoolID     *string    `json:"school_id"`
	PhotoURL     *string    `json:"photo_url"`
	Status       string     `json:"status"`
	AuthProvider string     `json:"auth_provider"`
	OTPEnabled   bool       `json:"otp_enabled"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	// student-only
	NIS                *string    `json:"nis"`
	UnlistedSchoolName *string    `json:"unlisted_school_name"`
	DOB                *time.Time `json:"dob"`
	Gender         *string    `json:"gender"`
	Grade          *int       `json:"grade"`
	AlamatDomisili *string    `json:"alamat_domisili"`
	TargetExam     *string    `json:"target_exam"`
}
