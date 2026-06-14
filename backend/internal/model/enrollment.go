package model

import (
	"time"

	"github.com/google/uuid"
)

type CourseEnrollment struct {
	ID         uuid.UUID
	StudentID  uuid.UUID
	ProductID  uuid.UUID
	OrderID    *uuid.UUID
	Status     string
	Source     string
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
