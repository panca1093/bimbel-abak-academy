package model

import (
	"time"

	"github.com/google/uuid"
)

type CourseSession struct {
	ID               uuid.UUID
	StudentID        uuid.UUID
	CourseID         uuid.UUID
	OrderID          *uuid.UUID
	Status           string
	Source           string
	EnrolledAt       time.Time
	RevokedAt        *time.Time
	CompletedLessons map[uuid.UUID]time.Time
}
