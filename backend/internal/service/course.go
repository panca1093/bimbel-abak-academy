package service

import (
	"context"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
)

// --- Course CRUD ---

func (s *Service) CreateCourse(ctx context.Context, title, level, subject, instructorName, role string) (model.Course, error) {
	if role != RoleAdminStore {
		return model.Course{}, ErrForbidden
	}

	c := model.Course{
		Title:          title,
		Level:          level,
		Subject:        subject,
		InstructorName: instructorName,
	}
	return s.storeRepo.CreateCourse(ctx, c)
}

func (s *Service) ListCourses(ctx context.Context, role string) ([]model.Course, error) {
	return s.storeRepo.ListCourses(ctx)
}

func (s *Service) UpdateCourse(ctx context.Context, id, title, level, subject, instructorName, role string) (model.Course, error) {
	if role != RoleAdminStore {
		return model.Course{}, ErrForbidden
	}

	courseID, err := parseUUID(id)
	if err != nil {
		return model.Course{}, err
	}

	c := model.Course{
		Title:          title,
		Level:          level,
		Subject:        subject,
		InstructorName: instructorName,
	}
	return s.storeRepo.UpdateCourse(ctx, courseID, c)
}

// --- Section CRUD (re-keyed to course_id) ---

func (s *Service) ListSections(ctx context.Context, courseID string) ([]model.Section, error) {
	cID, err := parseUUID(courseID)
	if err != nil {
		return nil, err
	}
	return s.storeRepo.ListSections(ctx, cID)
}

func (s *Service) CreateSection(ctx context.Context, courseID string, title string, role string) (model.Section, error) {
	if role != RoleAdminStore {
		return model.Section{}, ErrForbidden
	}

	cID, err := parseUUID(courseID)
	if err != nil {
		return model.Section{}, err
	}

	sections, err := s.storeRepo.ListSections(ctx, cID)
	if err != nil {
		return model.Section{}, err
	}

	position := len(sections)
	sec := model.Section{
		CourseID: cID,
		Title:    title,
		Position: position,
	}
	return s.storeRepo.CreateSection(ctx, sec)
}

func (s *Service) UpdateSection(ctx context.Context, courseID, sectionID string, title string, role string) (model.Section, error) {
	if role != RoleAdminStore {
		return model.Section{}, ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return model.Section{}, err
	}

	return s.storeRepo.UpdateSection(ctx, sID, title)
}

func (s *Service) DeleteSection(ctx context.Context, courseID, sectionID string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return err
	}

	return s.storeRepo.DeleteSection(ctx, sID)
}

func (s *Service) ReorderSections(ctx context.Context, courseID string, orderedIDs []string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	cID, err := parseUUID(courseID)
	if err != nil {
		return err
	}

	var ids []uuid.UUID
	for _, id := range orderedIDs {
		parsed, err := parseUUID(id)
		if err != nil {
			return err
		}
		ids = append(ids, parsed)
	}

	return s.storeRepo.ReorderSections(ctx, cID, ids)
}

// --- Lesson CRUD (re-keyed to course_id, unchanged child of section) ---

func (s *Service) CreateLesson(ctx context.Context, courseID, sectionID string, title, videoURL string, duration int, role string) (model.Lesson, error) {
	if role != RoleAdminStore {
		return model.Lesson{}, ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return model.Lesson{}, err
	}

	lessons, err := s.listLessonsBySection(ctx, sID)
	if err != nil {
		return model.Lesson{}, err
	}

	position := len(lessons)
	lesson := model.Lesson{
		SectionID:       sID,
		Title:           title,
		VideoURL:        videoURL,
		DurationSeconds: duration,
		Position:        position,
	}
	return s.storeRepo.CreateLesson(ctx, lesson)
}

func (s *Service) UpdateLesson(ctx context.Context, courseID, sectionID, lessonID string, title, videoURL string, duration int, role string) (model.Lesson, error) {
	if role != RoleAdminStore {
		return model.Lesson{}, ErrForbidden
	}

	lID, err := parseUUID(lessonID)
	if err != nil {
		return model.Lesson{}, err
	}

	lesson := model.Lesson{
		Title:           title,
		VideoURL:        videoURL,
		DurationSeconds: duration,
	}
	return s.storeRepo.UpdateLesson(ctx, lID, lesson)
}

func (s *Service) DeleteLesson(ctx context.Context, courseID, sectionID, lessonID string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	lID, err := parseUUID(lessonID)
	if err != nil {
		return err
	}

	return s.storeRepo.DeleteLesson(ctx, lID)
}

func (s *Service) ReorderLessons(ctx context.Context, courseID, sectionID string, orderedIDs []string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return err
	}

	var ids []uuid.UUID
	for _, id := range orderedIDs {
		parsed, err := parseUUID(id)
		if err != nil {
			return err
		}
		ids = append(ids, parsed)
	}

	return s.storeRepo.ReorderLessons(ctx, sID, ids)
}

func (s *Service) listLessonsBySection(ctx context.Context, sectionID uuid.UUID) ([]model.Lesson, error) {
	return s.storeRepo.ListLessonsBySection(ctx, sectionID)
}
