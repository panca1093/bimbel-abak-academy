package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
)

func TestCourseMethods(t *testing.T) {
	r := &Repository{}
	ctx := context.Background()

	_ = r.ListSections
	_ = r.CreateSection
	_ = r.UpdateSection
	_ = r.DeleteSection
	_ = r.ReorderSections

	_ = r.CreateLesson
	_ = r.UpdateLesson
	_ = r.DeleteLesson
	_ = r.ReorderLessons

	section := model.CourseSection{
		ID:        uuid.New(),
		ProductID: uuid.New(),
		Title:     "Section 1",
		Position:  0,
		CreatedAt: time.Now(),
	}
	_ = section

	lesson := model.Lesson{
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
