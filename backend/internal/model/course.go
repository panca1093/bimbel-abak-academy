package model

import (
	"time"

	"github.com/google/uuid"
)

type Course struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Level          string    `json:"level"`
	Subject        string    `json:"subject"`
	InstructorName string    `json:"instructor_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Section struct {
	ID        uuid.UUID `json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`
	Lessons   []Lesson  `json:"lessons"`
	CreatedAt time.Time `json:"created_at"`
}

type Lesson struct {
	ID              uuid.UUID `json:"id"`
	SectionID       uuid.UUID `json:"section_id"`
	Title           string    `json:"title"`
	VideoURL        string    `json:"video_url"`
	DurationSeconds int       `json:"duration_seconds"`
	Position        int       `json:"position"`
	CreatedAt       time.Time `json:"created_at"`
}

type CourseSession struct {
	ID               uuid.UUID            `json:"id"`
	StudentID        uuid.UUID            `json:"student_id"`
	CourseID         uuid.UUID            `json:"course_id"`
	OrderID          *uuid.UUID           `json:"order_id"`
	Status           string               `json:"status"`
	Source           string               `json:"source"`
	EnrolledAt       time.Time            `json:"enrolled_at"`
	RevokedAt        *time.Time           `json:"revoked_at"`
	CompletedLessons map[uuid.UUID]time.Time `json:"completed_lessons"`
}
