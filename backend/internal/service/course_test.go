package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"akademi-bimbel/internal/repository"
)

// fakeCourseRepo stubs course repository methods.
type fakeCourseRepo struct {
	sections map[string]*repository.CourseSection
	lessons  map[string]*repository.Lesson
	products map[string]*repository.Product
	seqSec   int
	seqLes   int
}

func newFakeCourseRepo() *fakeCourseRepo {
	return &fakeCourseRepo{
		sections: make(map[string]*repository.CourseSection),
		lessons:  make(map[string]*repository.Lesson),
		products: make(map[string]*repository.Product),
	}
}

func (f *fakeCourseRepo) ListSections(_ context.Context, productID uuid.UUID) ([]repository.CourseSection, error) {
	var result []repository.CourseSection
	for _, sec := range f.sections {
		if sec.ProductID == productID {
			result = append(result, *sec)
		}
	}
	return result, nil
}

func (f *fakeCourseRepo) CreateSection(_ context.Context, s repository.CourseSection) (repository.CourseSection, error) {
	s.ID = uuid.New()
	f.sections[s.ID.String()] = &s
	return s, nil
}

func (f *fakeCourseRepo) UpdateSection(_ context.Context, id uuid.UUID, title string) (repository.CourseSection, error) {
	sec, ok := f.sections[id.String()]
	if !ok {
		return repository.CourseSection{}, repository.ErrNotFound
	}
	sec.Title = title
	return *sec, nil
}

func (f *fakeCourseRepo) DeleteSection(_ context.Context, id uuid.UUID) error {
	delete(f.sections, id.String())
	return nil
}

func (f *fakeCourseRepo) ReorderSections(_ context.Context, productID uuid.UUID, orderedIDs []uuid.UUID) error {
	for i, id := range orderedIDs {
		sec, ok := f.sections[id.String()]
		if !ok || sec.ProductID != productID {
			return repository.ErrNotFound
		}
		sec.Position = i
	}
	return nil
}

func (f *fakeCourseRepo) CreateLesson(_ context.Context, l repository.Lesson) (repository.Lesson, error) {
	l.ID = uuid.New()
	f.lessons[l.ID.String()] = &l
	return l, nil
}

func (f *fakeCourseRepo) UpdateLesson(_ context.Context, id uuid.UUID, l repository.Lesson) (repository.Lesson, error) {
	lesson, ok := f.lessons[id.String()]
	if !ok {
		return repository.Lesson{}, repository.ErrNotFound
	}
	lesson.Title = l.Title
	lesson.VideoURL = l.VideoURL
	lesson.DurationSeconds = l.DurationSeconds
	return *lesson, nil
}

func (f *fakeCourseRepo) DeleteLesson(_ context.Context, id uuid.UUID) error {
	delete(f.lessons, id.String())
	return nil
}

func (f *fakeCourseRepo) ReorderLessons(_ context.Context, sectionID uuid.UUID, orderedIDs []uuid.UUID) error {
	for i, id := range orderedIDs {
		lesson, ok := f.lessons[id.String()]
		if !ok || lesson.SectionID != sectionID {
			return repository.ErrNotFound
		}
		lesson.Position = i
	}
	return nil
}

func (f *fakeCourseRepo) GetProductByID(_ context.Context, id string) (*repository.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakeCourseRepo) ListLessonsBySection(_ context.Context, sectionID uuid.UUID) ([]repository.Lesson, error) {
	var result []repository.Lesson
	for _, lesson := range f.lessons {
		if lesson.SectionID == sectionID {
			result = append(result, *lesson)
		}
	}
	return result, nil
}

func (f *fakeCourseRepo) seedProduct(p repository.Product) {
	f.products[p.ID] = &p
}

// shimCourseService implements course methods via fake repo.
type shimCourseService struct {
	fake *fakeCourseRepo
}

func (s *shimCourseService) ListSections(ctx context.Context, productID string) ([]repository.CourseSection, error) {
	pID, err := parseUUID(productID)
	if err != nil {
		return nil, err
	}

	product, err := s.fake.GetProductByID(ctx, pID.String())
	if err != nil {
		return nil, ErrProductNotFound
	}
	if product.Type != "course" {
		return nil, errors.New("product is not a course")
	}

	return s.fake.ListSections(ctx, pID)
}

func (s *shimCourseService) CreateSection(ctx context.Context, productID string, title string, role string) (repository.CourseSection, error) {
	if role != RoleAdminStore {
		return repository.CourseSection{}, ErrForbidden
	}

	pID, err := parseUUID(productID)
	if err != nil {
		return repository.CourseSection{}, err
	}

	product, err := s.fake.GetProductByID(ctx, pID.String())
	if err != nil {
		return repository.CourseSection{}, ErrProductNotFound
	}
	if product.Type != "course" {
		return repository.CourseSection{}, errors.New("product is not a course")
	}

	sections, err := s.fake.ListSections(ctx, pID)
	if err != nil {
		return repository.CourseSection{}, err
	}

	position := len(sections)
	sec := repository.CourseSection{
		ProductID: pID,
		Title:     title,
		Position:  position,
	}
	return s.fake.CreateSection(ctx, sec)
}

func (s *shimCourseService) UpdateSection(ctx context.Context, productID, sectionID string, title string, role string) (repository.CourseSection, error) {
	if role != RoleAdminStore {
		return repository.CourseSection{}, ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return repository.CourseSection{}, err
	}

	return s.fake.UpdateSection(ctx, sID, title)
}

func (s *shimCourseService) DeleteSection(ctx context.Context, productID, sectionID string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return err
	}

	return s.fake.DeleteSection(ctx, sID)
}

func (s *shimCourseService) ReorderSections(ctx context.Context, productID string, orderedIDs []string, role string) error {
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

	return s.fake.ReorderSections(ctx, pID, ids)
}

func (s *shimCourseService) CreateLesson(ctx context.Context, productID, sectionID string, title, videoURL string, duration int, role string) (repository.Lesson, error) {
	if role != RoleAdminStore {
		return repository.Lesson{}, ErrForbidden
	}

	sID, err := parseUUID(sectionID)
	if err != nil {
		return repository.Lesson{}, err
	}

	lessons, err := s.fake.ListLessonsBySection(ctx, sID)
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
	return s.fake.CreateLesson(ctx, lesson)
}

func (s *shimCourseService) UpdateLesson(ctx context.Context, productID, sectionID, lessonID string, title, videoURL string, duration int, role string) (repository.Lesson, error) {
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
	return s.fake.UpdateLesson(ctx, lID, lesson)
}

func (s *shimCourseService) DeleteLesson(ctx context.Context, productID, sectionID, lessonID string, role string) error {
	if role != RoleAdminStore {
		return ErrForbidden
	}

	lID, err := parseUUID(lessonID)
	if err != nil {
		return err
	}

	return s.fake.DeleteLesson(ctx, lID)
}

func (s *shimCourseService) ReorderLessons(ctx context.Context, productID, sectionID string, orderedIDs []string, role string) error {
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

	return s.fake.ReorderLessons(ctx, sID, ids)
}

// Test: ListSections validates product exists and is type course
func TestListSections_ValidatesCourseType(t *testing.T) {
	ctx := context.Background()
	fake := newFakeCourseRepo()
	svc := &shimCourseService{fake: fake}

	bookID := uuid.New().String()
	bookProd := repository.Product{ID: bookID, Type: "book"}
	fake.seedProduct(bookProd)

	_, err := svc.ListSections(ctx, bookID)
	if err == nil || err.Error() != "product is not a course" {
		t.Fatalf("expected product type error, got %v", err)
	}
}

// Test: CreateSection rejects non-course products
func TestCreateSection_RejectsNonCourse(t *testing.T) {
	ctx := context.Background()
	fake := newFakeCourseRepo()
	svc := &shimCourseService{fake: fake}

	bookID := uuid.New().String()
	bookProd := repository.Product{ID: bookID, Type: "book"}
	fake.seedProduct(bookProd)

	_, err := svc.CreateSection(ctx, bookID, "Test Section", RoleAdminStore)
	if err == nil || err.Error() != "product is not a course" {
		t.Fatalf("expected product type error, got %v", err)
	}
}

// Test: ReorderSections rejects mismatched IDs
func TestReorderSections_RejectsMismatchedIDs(t *testing.T) {
	ctx := context.Background()
	fake := newFakeCourseRepo()
	svc := &shimCourseService{fake: fake}

	prodID := uuid.New()
	courseProd := repository.Product{ID: prodID.String(), Type: "course"}
	fake.seedProduct(courseProd)

	sec1 := repository.CourseSection{ID: uuid.New(), ProductID: prodID, Title: "Sec 1", Position: 0}
	fake.sections[sec1.ID.String()] = &sec1

	wrongSecID := uuid.New()
	err := svc.ReorderSections(ctx, prodID.String(), []string{wrongSecID.String()}, RoleAdminStore)
	if err == nil || err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound for mismatched section, got %v", err)
	}
}
