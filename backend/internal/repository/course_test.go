package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCourseMethods(t *testing.T) {
	// Compile-time test: verify all Course and Lesson methods exist on *Repository

	r := &Repository{}
	ctx := context.Background()

	// CourseSection methods
	_ = r.ListSections
	_ = r.CreateSection
	_ = r.UpdateSection
	_ = r.DeleteSection
	_ = r.ReorderSections

	// Lesson methods
	_ = r.CreateLesson
	_ = r.UpdateLesson
	_ = r.DeleteLesson
	_ = r.ReorderLessons

	// Verify struct field names compile
	section := CourseSection{
		ID:        uuid.New(),
		ProductID: uuid.New(),
		Title:     "Section 1",
		Position:  0,
		CreatedAt: time.Now(),
	}
	_ = section

	lesson := Lesson{
		ID:              uuid.New(),
		SectionID:       uuid.New(),
		Title:           "Lesson 1",
		VideoURL:        "https://example.com/video",
		DurationSeconds: 300,
		Position:        0,
		CreatedAt:       time.Now(),
	}
	_ = lesson

	_ = ctx
}
