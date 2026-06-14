package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"akademi-bimbel/internal/repository"
)

var ErrNotCourse = errors.New("product is not a course")

func (s *Service) ListSections(ctx context.Context, productID string) ([]repository.CourseSection, error) {
	pID, err := parseUUID(productID)
	if err != nil {
		return nil, err
	}

	product, err := s.storeRepo.GetProductByID(ctx, pID.String())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if product.Type != "course" {
		return nil, ErrNotCourse
	}

	return s.storeRepo.ListSections(ctx, pID)
}

func (s *Service) CreateSection(ctx context.Context, productID string, title string, role string) (repository.CourseSection, error) {
	if role != RoleAdminStore {
		return repository.CourseSection{}, ErrForbidden
	}

	pID, err := parseUUID(productID)
	if err != nil {
		return repository.CourseSection{}, err
	}

	product, err := s.storeRepo.GetProductByID(ctx, pID.String())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.CourseSection{}, ErrProductNotFound
		}
		return repository.CourseSection{}, err
	}
	if product.Type != "course" {
		return repository.CourseSection{}, ErrNotCourse
	}

	sections, err := s.storeRepo.ListSections(ctx, pID)
	if err != nil {
		return repository.CourseSection{}, err
	}

	position := len(sections)
	sec := repository.CourseSection{
		ProductID: pID,
		Title:     title,
		Position:  position,
	}
	return s.storeRepo.CreateSection(ctx, sec)
}

func (s *Service) UpdateSection(ctx context.Context, productID, sectionID string, title string, role string) (repository.CourseSection, error) {
	if role != RoleAdminStore {
		return repository.CourseSection{}, ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return repository.CourseSection{}, err
	}

	return s.storeRepo.UpdateSection(ctx, sID, title)
}

func (s *Service) DeleteSection(ctx context.Context, productID, sectionID string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return err
	}

	return s.storeRepo.DeleteSection(ctx, sID)
}

func (s *Service) ReorderSections(ctx context.Context, productID string, orderedIDs []string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	pID, err := parseUUID(productID)
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

	return s.storeRepo.ReorderSections(ctx, pID, ids)
}

func (s *Service) CreateLesson(ctx context.Context, productID, sectionID string, title, videoURL string, duration int, role string) (repository.Lesson, error) {
	if role != RoleAdminStore {
		return repository.Lesson{}, ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return repository.Lesson{}, err
	}

	lessons, err := s.listLessonsBySection(ctx, sID)
	if err != nil {
		return repository.Lesson{}, err
	}

	position := len(lessons)
	lesson := repository.Lesson{
		SectionID:       sID,
		Title:           title,
		VideoURL:        videoURL,
		DurationSeconds: duration,
		Position:        position,
	}
	return s.storeRepo.CreateLesson(ctx, lesson)
}

func (s *Service) UpdateLesson(ctx context.Context, productID, sectionID, lessonID string, title, videoURL string, duration int, role string) (repository.Lesson, error) {
	if role != RoleAdminStore {
		return repository.Lesson{}, ErrForbidden
	}

	lID, err := parseUUID(lessonID)
	if err != nil {
		return repository.Lesson{}, err
	}

	lesson := repository.Lesson{
		Title:           title,
		VideoURL:        videoURL,
		DurationSeconds: duration,
	}
	return s.storeRepo.UpdateLesson(ctx, lID, lesson)
}

func (s *Service) DeleteLesson(ctx context.Context, productID, sectionID, lessonID string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	lID, err := parseUUID(lessonID)
	if err != nil {
		return err
	}

	return s.storeRepo.DeleteLesson(ctx, lID)
}

func (s *Service) ReorderLessons(ctx context.Context, productID, sectionID string, orderedIDs []string, role string) error {
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

func (s *Service) listLessonsBySection(ctx context.Context, sectionID uuid.UUID) ([]repository.Lesson, error) {
	return s.storeRepo.ListLessonsBySection(ctx, sectionID)
}
