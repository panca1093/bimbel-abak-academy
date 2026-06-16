package model

import (
	"time"

	"github.com/google/uuid"
)

type Course struct {
	ID             uuid.UUID
	Title          string
	Level          string
	Subject        string
	InstructorName string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Section struct {
	ID        uuid.UUID
	CourseID  uuid.UUID
	Title     string
	Position  int
	CreatedAt time.Time
}

type Lesson struct {
	ID              uuid.UUID
	SectionID       uuid.UUID
	Title           string
	VideoURL        string
	DurationSeconds int
	Position        int
	CreatedAt       time.Time
}
