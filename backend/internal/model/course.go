package model

import (
	"time"

	"github.com/google/uuid"
)

type CourseSection struct {
	ID        uuid.UUID
	ProductID uuid.UUID
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
