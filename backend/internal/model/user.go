package model

import "time"

type User struct {
	ID           string
	Email        *string
	Username     *string
	Phone        *string
	PasswordHash string
	Role         string
	Name         string
	SchoolID     *string
	Status       string
	OTPEnabled   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// student-only
	NIS            *string
	DOB            *time.Time
	Gender         *string
	Grade          *int
	AlamatDomisili *string
	TargetExam     *string
}
