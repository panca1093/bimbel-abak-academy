package service

import (
	"context"
	"errors"
	"testing"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

// findSchool pages through AdminListSchools looking for a school by ID. The
// real-DB fixture is shared across every test in this package, so a single
// page isn't guaranteed to contain a school created by this test.
func findSchool(t *testing.T, svc *Service, id string) SchoolResponse {
	t.Helper()
	ctx := context.Background()
	cursor := ""
	for {
		rows, next, err := svc.AdminListSchools(ctx, 100, cursor)
		if err != nil {
			t.Fatalf("AdminListSchools: %v", err)
		}
		for _, r := range rows {
			if r.ID == id {
				return r
			}
		}
		if next == "" {
			t.Fatalf("school %s not found in AdminListSchools", id)
		}
		cursor = next
	}
}

func TestCreateSchool_Integration(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	t.Run("happy path creates active school with zero student count", func(t *testing.T) {
		code := "cs_" + uniqueSuffix()
		npsn := "20000001"
		alamat := "Jl. Test No.1"
		resp, err := svc.CreateSchool(ctx, "Test School "+code, code, &npsn, []string{"SMA"}, &alamat)
		if err != nil {
			t.Fatalf("CreateSchool: %v", err)
		}
		if resp.Status != "active" {
			t.Errorf("Status: want active, got %q", resp.Status)
		}
		if resp.StudentCount != 0 {
			t.Errorf("StudentCount: want 0, got %d", resp.StudentCount)
		}
		if resp.ID == "" {
			t.Error("want non-empty ID")
		}
		if resp.Code != code {
			t.Errorf("Code: want %s, got %s", code, resp.Code)
		}
	})

	t.Run("omitted school_types defaults to empty slice not null", func(t *testing.T) {
		code := "cs_" + uniqueSuffix()
		resp, err := svc.CreateSchool(ctx, "No Types School "+code, code, nil, nil, nil)
		if err != nil {
			t.Fatalf("CreateSchool with omitted school_types: %v", err)
		}
		if resp.SchoolTypes == nil {
			t.Error("SchoolTypes: want empty slice, got nil")
		}
		if len(resp.SchoolTypes) != 0 {
			t.Errorf("SchoolTypes: want empty, got %v", resp.SchoolTypes)
		}

		// Persisted row must also carry {} not NULL, and remain listable.
		found := findSchool(t, svc, resp.ID)
		if found.SchoolTypes == nil || len(found.SchoolTypes) != 0 {
			t.Errorf("persisted SchoolTypes: want empty slice, got %v", found.SchoolTypes)
		}
	})

	t.Run("missing name returns ErrMissingField", func(t *testing.T) {
		_, err := svc.CreateSchool(ctx, "", "somecode", nil, nil, nil)
		if !errors.Is(err, ErrMissingField) {
			t.Errorf("want ErrMissingField, got %v", err)
		}
	})

	t.Run("missing code returns ErrMissingField", func(t *testing.T) {
		_, err := svc.CreateSchool(ctx, "Some School", "", nil, nil, nil)
		if !errors.Is(err, ErrMissingField) {
			t.Errorf("want ErrMissingField, got %v", err)
		}
	})

	t.Run("duplicate code returns ErrSchoolCodeTaken", func(t *testing.T) {
		code := "cs_" + uniqueSuffix()
		if _, err := svc.CreateSchool(ctx, "First", code, nil, nil, nil); err != nil {
			t.Fatalf("CreateSchool (first): %v", err)
		}
		_, err := svc.CreateSchool(ctx, "Second", code, nil, nil, nil)
		if !errors.Is(err, ErrSchoolCodeTaken) {
			t.Errorf("want ErrSchoolCodeTaken, got %v", err)
		}
	})
}

func TestUpdateSchool_Integration(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	t.Run("happy path patches fields", func(t *testing.T) {
		code := "us_" + uniqueSuffix()
		created, err := svc.CreateSchool(ctx, "Before Update", code, nil, nil, nil)
		if err != nil {
			t.Fatalf("CreateSchool: %v", err)
		}
		newName := "After Update"
		updated, err := svc.UpdateSchool(ctx, created.ID, &newName, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("UpdateSchool: %v", err)
		}
		if updated.Name != newName {
			t.Errorf("Name: want %s, got %s", newName, updated.Name)
		}
		if updated.Code != code {
			t.Errorf("Code should be unchanged: want %s, got %s", code, updated.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		newName := "Doesn't Matter"
		_, err := svc.UpdateSchool(ctx, "00000000-0000-0000-0000-000000000000", &newName, nil, nil, nil, nil)
		if !errors.Is(err, ErrSchoolNotFound) {
			t.Errorf("want ErrSchoolNotFound, got %v", err)
		}
	})

	t.Run("code uniqueness on update", func(t *testing.T) {
		codeA := "us_" + uniqueSuffix()
		codeB := "us_" + uniqueSuffix()
		if _, err := svc.CreateSchool(ctx, "School A", codeA, nil, nil, nil); err != nil {
			t.Fatalf("CreateSchool A: %v", err)
		}
		schoolB, err := svc.CreateSchool(ctx, "School B", codeB, nil, nil, nil)
		if err != nil {
			t.Fatalf("CreateSchool B: %v", err)
		}
		_, err = svc.UpdateSchool(ctx, schoolB.ID, nil, nil, nil, nil, &codeA)
		if !errors.Is(err, ErrSchoolCodeTaken) {
			t.Errorf("want ErrSchoolCodeTaken, got %v", err)
		}
	})

	t.Run("code change succeeds when students exist (lock removed)", func(t *testing.T) {
		code := "us_" + uniqueSuffix()
		school, err := svc.CreateSchool(ctx, "School With Students", code, nil, nil, nil)
		if err != nil {
			t.Fatalf("CreateSchool: %v", err)
		}
		if _, err := svc.RegisterStudent(ctx, school.ID, "Stu Dent", "sma", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
			t.Fatalf("RegisterStudent: %v", err)
		}
		newCode := "us_" + uniqueSuffix()
		updated, err := svc.UpdateSchool(ctx, school.ID, nil, nil, nil, nil, &newCode)
		if err != nil {
			t.Errorf("code change should succeed (lock removed), got %v", err)
		}
		if updated != nil && updated.Code != newCode {
			t.Errorf("Code: want %s, got %s", newCode, updated.Code)
		}
	})

	t.Run("code uniqueness still enforced on update", func(t *testing.T) {
		codeA := "us_" + uniqueSuffix()
		codeB := "us_" + uniqueSuffix()
		_, err := svc.CreateSchool(ctx, "School A", codeA, nil, nil, nil)
		if err != nil {
			t.Fatalf("CreateSchool A: %v", err)
		}
		schoolB, err := svc.CreateSchool(ctx, "School B", codeB, nil, nil, nil)
		if err != nil {
			t.Fatalf("CreateSchool B: %v", err)
		}
		if _, err := svc.RegisterStudent(ctx, schoolB.ID, "Stu Dent", "sma", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
			t.Fatalf("RegisterStudent: %v", err)
		}
		_, err = svc.UpdateSchool(ctx, schoolB.ID, nil, nil, nil, nil, &codeA)
		if !errors.Is(err, ErrSchoolCodeTaken) {
			t.Errorf("want ErrSchoolCodeTaken, got %v", err)
		}
	})
}

func TestChangeSchoolStatus_Integration(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	t.Run("happy path toggles status", func(t *testing.T) {
		code := "st_" + uniqueSuffix()
		school, err := svc.CreateSchool(ctx, "Status School", code, nil, nil, nil)
		if err != nil {
			t.Fatalf("CreateSchool: %v", err)
		}
		updated, err := svc.ChangeSchoolStatus(ctx, school.ID, "deactivated")
		if err != nil {
			t.Fatalf("ChangeSchoolStatus: %v", err)
		}
		if updated.Status != "deactivated" {
			t.Errorf("Status: want deactivated, got %s", updated.Status)
		}
		updated, err = svc.ChangeSchoolStatus(ctx, school.ID, "active")
		if err != nil {
			t.Fatalf("ChangeSchoolStatus back to active: %v", err)
		}
		if updated.Status != "active" {
			t.Errorf("Status: want active, got %s", updated.Status)
		}
	})

	t.Run("invalid status value", func(t *testing.T) {
		code := "st_" + uniqueSuffix()
		school, err := svc.CreateSchool(ctx, "Invalid Status School", code, nil, nil, nil)
		if err != nil {
			t.Fatalf("CreateSchool: %v", err)
		}
		_, err = svc.ChangeSchoolStatus(ctx, school.ID, "pending")
		if !errors.Is(err, ErrInvalidStatusFilter) {
			t.Errorf("want ErrInvalidStatusFilter, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.ChangeSchoolStatus(ctx, "00000000-0000-0000-0000-000000000000", "active")
		if !errors.Is(err, ErrSchoolNotFound) {
			t.Errorf("want ErrSchoolNotFound, got %v", err)
		}
	})
}

func TestAdminListSchools_Integration(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	code := "ls_" + uniqueSuffix()
	name := "Listable School " + code
	school, err := svc.CreateSchool(ctx, name, code, nil, nil, nil)
	if err != nil {
		t.Fatalf("CreateSchool: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := svc.RegisterStudent(ctx, school.ID, "Stu Dent", "sma", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
			t.Fatalf("RegisterStudent: %v", err)
		}
	}

	found := findSchool(t, svc, school.ID)
	if found.Name != name {
		t.Errorf("Name: want %s, got %s", name, found.Name)
	}
	if found.StudentCount != 2 {
		t.Errorf("StudentCount: want 2, got %d", found.StudentCount)
	}
}

func TestSchoolSentinelErrors(t *testing.T) {
	if ErrSchoolNotFound == nil {
		t.Error("ErrSchoolNotFound is nil")
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
