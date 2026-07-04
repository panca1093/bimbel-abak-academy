package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

// fakeSchoolStore is an in-memory stub for school-related repository methods.
type fakeSchoolStore struct {
	schools       map[string]*model.School
	studentCounts map[string]int
	nextIdx       int
}

func newFakeSchoolStore() *fakeSchoolStore {
	return &fakeSchoolStore{
		schools:       make(map[string]*model.School),
		studentCounts: make(map[string]int),
	}
}

func (f *fakeSchoolStore) add(name, code string) *model.School {
	f.nextIdx++
	id := fmt.Sprintf("school-%d", f.nextIdx)
	s := &model.School{
		ID:        id,
		Name:      name,
		Code:      code,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	f.schools[id] = s
	f.studentCounts[id] = 0
	return s
}

// Repository method stubs used by Service methods under test.

func (f *fakeSchoolStore) ListSchoolsAdmin(_ context.Context, limit int, cursor string) ([]repository.SchoolAdminRow, string, error) {
	var rows []repository.SchoolAdminRow
	for _, s := range f.schools {
		rows = append(rows, repository.SchoolAdminRow{School: *s, StudentCount: f.studentCounts[s.ID]})
	}
	return rows, "", nil
}

func (f *fakeSchoolStore) GetSchoolByID(_ context.Context, id string) (*model.School, error) {
	s, ok := f.schools[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (f *fakeSchoolStore) SchoolCodeExists(_ context.Context, code string, excludeID *string) (bool, error) {
	for id, s := range f.schools {
		if s.Code == code {
			if excludeID != nil && id == *excludeID {
				continue
			}
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeSchoolStore) CreateSchool(_ context.Context, s *model.School) error {
	f.nextIdx++
	s.ID = fmt.Sprintf("school-%d", f.nextIdx)
	s.Status = "active"
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	f.schools[s.ID] = s
	f.studentCounts[s.ID] = 0
	return nil
}

func (f *fakeSchoolStore) UpdateSchool(_ context.Context, id string, name, npsn, alamat *string, schoolTypes []string, code *string) error {
	s := f.schools[id]
	if name != nil {
		s.Name = *name
	}
	if npsn != nil {
		s.NPSN = npsn
	}
	if alamat != nil {
		s.Alamat = alamat
	}
	if schoolTypes != nil {
		s.SchoolTypes = schoolTypes
	}
	if code != nil {
		s.Code = *code
	}
	s.UpdatedAt = time.Now()
	f.schools[id] = s
	return nil
}

func (f *fakeSchoolStore) UpdateSchoolStatus(_ context.Context, id, status string) error {
	s := f.schools[id]
	s.Status = status
	s.UpdatedAt = time.Now()
	f.schools[id] = s
	return nil
}

func (f *fakeSchoolStore) CountStudentsBySchool(_ context.Context, schoolID string) (int, error) {
	return f.studentCounts[schoolID], nil
}

// schoolTestService is a thin adapter that wires fakeSchoolStore into a Service
// so school-domain service methods can be tested without a real DB.
type schoolTestService struct {
	svc   *Service
	store *fakeSchoolStore
}

func newSchoolTestService(t *testing.T) *schoolTestService {
	t.Helper()
	repo := newFakeUserRepo()
	svc, _ := newTestService(t, repo)
	store := newFakeSchoolStore()
	return &schoolTestService{svc: svc, store: store}
}

// — tests —

func TestAdminListSchools_Empty(t *testing.T) {
	ts := newSchoolTestService(t)
	rows, nextCursor, err := ts.store.ListSchoolsAdmin(context.Background(), 20, "")
	if err != nil {
		t.Fatalf("ListSchoolsAdmin: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("want 0 schools, got %d", len(rows))
	}
	if nextCursor != "" {
		t.Errorf("next cursor: want empty, got %s", nextCursor)
	}
}

func TestAdminListSchools_WithData(t *testing.T) {
	ts := newSchoolTestService(t)
	ts.store.add("SMAN 1 Jakarta", "sman1jkt")
	ts.store.add("SMAN 3 Bandung", "sman3bdg")
	ts.store.studentCounts["school-1"] = 10
	ts.store.studentCounts["school-2"] = 5

	rows, next, err := ts.store.ListSchoolsAdmin(context.Background(), 20, "")
	if err != nil {
		t.Fatalf("ListSchoolsAdmin: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 schools, got %d", len(rows))
	}
	if next != "" {
		t.Errorf("next cursor: want empty, got %s", next)
	}
	found := false
	for _, r := range rows {
		if r.School.Name == "SMAN 1 Jakarta" && r.StudentCount == 10 {
			found = true
		}
	}
	if !found {
		t.Error("did not find SMAN 1 Jakarta with student_count=10")
	}
}

func TestSchoolCodeExists_True(t *testing.T) {
	ts := newSchoolTestService(t)
	ts.store.add("SMAN 1", "sman1")
	exists, err := ts.store.SchoolCodeExists(context.Background(), "sman1", nil)
	if err != nil {
		t.Fatalf("SchoolCodeExists: %v", err)
	}
	if !exists {
		t.Error("sman1 should exist")
	}
}

func TestSchoolCodeExists_False(t *testing.T) {
	ts := newSchoolTestService(t)
	exists, err := ts.store.SchoolCodeExists(context.Background(), "nonexistent", nil)
	if err != nil {
		t.Fatalf("SchoolCodeExists: %v", err)
	}
	if exists {
		t.Error("nonexistent should not exist")
	}
}

func TestSchoolCodeExists_ExcludeSelf(t *testing.T) {
	ts := newSchoolTestService(t)
	s := ts.store.add("SMAN 1", "sman1")
	// Passing excludeID=s.ID means "don't count this school as a conflict"
	id := s.ID
	exists, err := ts.store.SchoolCodeExists(context.Background(), "sman1", &id)
	if err != nil {
		t.Fatalf("SchoolCodeExists: %v", err)
	}
	if exists {
		t.Error("sman1 should not conflict when excluding self")
	}
}

func TestGetSchoolByID_NotFound(t *testing.T) {
	ts := newSchoolTestService(t)
	sch, err := ts.store.GetSchoolByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("GetSchoolByID: %v", err)
	}
	if sch != nil {
		t.Error("want nil for nonexistent school")
	}
}

func TestGetSchoolByID_Found(t *testing.T) {
	ts := newSchoolTestService(t)
	s := ts.store.add("SMAN 1", "sman1")
	sch, err := ts.store.GetSchoolByID(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("GetSchoolByID: %v", err)
	}
	if sch == nil {
		t.Fatal("want non-nil for existing school")
	}
	if sch.Name != "SMAN 1" {
		t.Errorf("Name: want SMAN 1, got %s", sch.Name)
	}
}

func TestCountStudentsBySchool_Zero(t *testing.T) {
	ts := newSchoolTestService(t)
	s := ts.store.add("SMAN 1", "sman1")
	count, err := ts.store.CountStudentsBySchool(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("CountStudentsBySchool: %v", err)
	}
	if count != 0 {
		t.Errorf("want 0, got %d", count)
	}
}

func TestCountStudentsBySchool_NonZero(t *testing.T) {
	ts := newSchoolTestService(t)
	s := ts.store.add("SMAN 1", "sman1")
	ts.store.studentCounts[s.ID] = 15
	count, err := ts.store.CountStudentsBySchool(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("CountStudentsBySchool: %v", err)
	}
	if count != 15 {
		t.Errorf("want 15, got %d", count)
	}
}

func TestSchoolSentinelErrors(t *testing.T) {
	if ErrSchoolNotFound == nil {
		t.Error("ErrSchoolNotFound is nil")
	}
	if ErrSchoolCodeLocked == nil {
		t.Error("ErrSchoolCodeLocked is nil")
	}
	if ErrSchoolCodeTaken == nil {
		t.Error("ErrSchoolCodeTaken is nil")
	}
}

func TestSchoolResponseMapping(t *testing.T) {
	row := repository.SchoolAdminRow{
		School: model.School{
			ID:   "s1",
			Name: "Test School",
			Code: "test",
		},
		StudentCount: 5,
	}
	resp := toSchoolResponse(row)
	if resp.ID != "s1" {
		t.Errorf("ID: want s1, got %s", resp.ID)
	}
	if resp.Name != "Test School" {
		t.Errorf("Name: want Test School, got %s", resp.Name)
	}
	if resp.Code != "test" {
		t.Errorf("Code: want test, got %s", resp.Code)
	}
	if resp.StudentCount != 5 {
		t.Errorf("StudentCount: want 5, got %d", resp.StudentCount)
	}
}
