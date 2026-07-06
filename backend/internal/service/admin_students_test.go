package service

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGenTempPassword(t *testing.T) {
	p1, err := genTempPassword()
	if err != nil {
		t.Fatalf("genTempPassword: %v", err)
	}
	if len(p1) != tempPasswordLen {
		t.Errorf("want length %d, got %d", tempPasswordLen, len(p1))
	}

	// Two calls should produce different passwords
	p2, err := genTempPassword()
	if err != nil {
		t.Fatalf("genTempPassword: %v", err)
	}
	if p1 == p2 {
		t.Error("two successive calls returned the same password")
	}
}

func TestGenTempPassword_Characters(t *testing.T) {
	p, err := genTempPassword()
	if err != nil {
		t.Fatalf("genTempPassword: %v", err)
	}
	for _, r := range p {
		found := false
		for _, c := range tempPasswordChars {
			if r == c {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("character %q not in allowed set", r)
		}
	}
}

func TestTempPasswordIsBcryptable(t *testing.T) {
	p, err := genTempPassword()
	if err != nil {
		t.Fatalf("genTempPassword: %v", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(p), 12)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte(p)); err != nil {
		t.Error("bcrypt compare failed for generated temp password")
	}
}

func TestStudentSentinelErrors(t *testing.T) {
	if ErrDuplicateNIS == nil {
		t.Error("ErrDuplicateNIS is nil")
	}
	if ErrSchoolDeactivated == nil {
		t.Error("ErrSchoolDeactivated is nil")
	}
	if ErrStudentNotFound == nil {
		t.Error("ErrStudentNotFound is nil")
	}
}

// createTestSchool is a small helper shared by the RegisterStudent/ListStudents
// integration tests below — it creates a school via the real Service so tests
// stay end-to-end rather than reaching around the Service into raw SQL.
func createTestSchool(t *testing.T, svc *Service) string {
	t.Helper()
	code := "stu_" + uniqueSuffix()
	resp, err := svc.CreateSchool(context.Background(), "Student Test School "+code, code, nil, nil, nil)
	if err != nil {
		t.Fatalf("CreateSchool: %v", err)
	}
	return resp.ID
}

func TestRegisterStudent_Integration(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	t.Run("happy path: username format, temp password once, bcrypt hash persisted", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		nis := "nis_" + uniqueSuffix()
		resp, err := svc.RegisterStudent(ctx, schoolID, "Budi Santoso", nis, nil, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("RegisterStudent: %v", err)
		}
		if resp.TempPassword == "" {
			t.Error("want non-empty temp_password")
		}
		if resp.NIS != nis {
			t.Errorf("NIS: want %s, got %s", nis, resp.NIS)
		}

		u, err := repo.GetUserByUsername(ctx, resp.Username)
		if err != nil {
			t.Fatalf("GetUserByUsername: %v", err)
		}
		if u == nil {
			t.Fatal("student user not persisted")
		}
		if u.Role != RoleStudent || u.Status != "active" || u.OTPEnabled {
			t.Errorf("unexpected defaults: role=%s status=%s otp=%v", u.Role, u.Status, u.OTPEnabled)
		}
		if u.PasswordHash == resp.TempPassword {
			t.Error("password hash must not equal the plaintext temp password")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(resp.TempPassword)); err != nil {
			t.Errorf("persisted hash does not match returned temp password: %v", err)
		}
	})

	t.Run("missing name returns ErrMissingField", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		_, err := svc.RegisterStudent(ctx, schoolID, "", "somenis", nil, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrMissingField) {
			t.Errorf("want ErrMissingField, got %v", err)
		}
	})

	t.Run("missing nis returns ErrMissingField", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		_, err := svc.RegisterStudent(ctx, schoolID, "Some Name", "", nil, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrMissingField) {
			t.Errorf("want ErrMissingField, got %v", err)
		}
	})

	t.Run("nonexistent school returns ErrSchoolNotFound", func(t *testing.T) {
		_, err := svc.RegisterStudent(ctx, "00000000-0000-0000-0000-000000000000", "Some Name", "somenis", nil, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrSchoolNotFound) {
			t.Errorf("want ErrSchoolNotFound, got %v", err)
		}
	})

	t.Run("duplicate NIS in same school returns clean ErrDuplicateNIS", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		nis := "nis_" + uniqueSuffix()
		if _, err := svc.RegisterStudent(ctx, schoolID, "First Student", nis, nil, nil, nil, nil, nil, nil); err != nil {
			t.Fatalf("RegisterStudent (first): %v", err)
		}
		_, err := svc.RegisterStudent(ctx, schoolID, "Second Student", nis, nil, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrDuplicateNIS) {
			t.Errorf("want ErrDuplicateNIS, got %v", err)
		}
	})

	t.Run("deactivated school blocks registration", func(t *testing.T) {
		schoolID := createTestSchool(t, svc)
		if _, err := svc.ChangeSchoolStatus(ctx, schoolID, "deactivated"); err != nil {
			t.Fatalf("ChangeSchoolStatus: %v", err)
		}
		_, err := svc.RegisterStudent(ctx, schoolID, "Blocked Student", "nis_"+uniqueSuffix(), nil, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrSchoolDeactivated) {
			t.Errorf("want ErrSchoolDeactivated, got %v", err)
		}
	})
}

func TestListStudents_ChangeStatus_Reissue_Integration(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	schoolA := createTestSchool(t, svc)
	schoolB := createTestSchool(t, svc)

	nis := "nis_" + uniqueSuffix()
	reg, err := svc.RegisterStudent(ctx, schoolA, "Row Scoped Student", nis, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("RegisterStudent: %v", err)
	}
	studentID := reg.ID

	t.Run("list is scoped to school", func(t *testing.T) {
		rowsA, _, err := svc.ListStudents(ctx, schoolA, "", "", 20, "")
		if err != nil {
			t.Fatalf("ListStudents (schoolA): %v", err)
		}
		foundInA := false
		for _, r := range rowsA {
			if r.ID == studentID {
				foundInA = true
			}
		}
		if !foundInA {
			t.Error("student should be listed under its own school")
		}

		rowsB, _, err := svc.ListStudents(ctx, schoolB, "", "", 20, "")
		if err != nil {
			t.Fatalf("ListStudents (schoolB): %v", err)
		}
		for _, r := range rowsB {
			if r.ID == studentID {
				t.Error("student from schoolA must not be listed under schoolB")
			}
		}
	})

	t.Run("change status is row-scoped: wrong school returns ErrStudentNotFound", func(t *testing.T) {
		err := svc.ChangeStudentStatus(ctx, schoolB, studentID, "deactivated")
		if !errors.Is(err, ErrStudentNotFound) {
			t.Errorf("want ErrStudentNotFound for cross-school access, got %v", err)
		}
	})

	t.Run("change status succeeds for the owning school", func(t *testing.T) {
		if err := svc.ChangeStudentStatus(ctx, schoolA, studentID, "deactivated"); err != nil {
			t.Fatalf("ChangeStudentStatus: %v", err)
		}
		rows, _, err := svc.ListStudents(ctx, schoolA, "", "", 20, "")
		if err != nil {
			t.Fatalf("ListStudents: %v", err)
		}
		var status string
		for _, r := range rows {
			if r.ID == studentID {
				status = r.Status
			}
		}
		if status != "deactivated" {
			t.Errorf("Status: want deactivated, got %s", status)
		}
		// restore to active for the reissue subtests below
		if err := svc.ChangeStudentStatus(ctx, schoolA, studentID, "active"); err != nil {
			t.Fatalf("ChangeStudentStatus (restore): %v", err)
		}
	})

	t.Run("credential reissue is row-scoped: wrong school returns ErrStudentNotFound", func(t *testing.T) {
		_, err := svc.ReissueStudentCredentials(ctx, schoolB, studentID)
		if !errors.Is(err, ErrStudentNotFound) {
			t.Errorf("want ErrStudentNotFound for cross-school access, got %v", err)
		}
	})

	t.Run("credential reissue overwrites hash and returns a new password", func(t *testing.T) {
		creds, err := svc.ReissueStudentCredentials(ctx, schoolA, studentID)
		if err != nil {
			t.Fatalf("ReissueStudentCredentials: %v", err)
		}
		if creds.TempPassword == "" {
			t.Fatal("want non-empty temp_password")
		}
		if creds.TempPassword == reg.TempPassword {
			t.Error("reissue should return a different password than the original registration")
		}

		u, err := repo.GetUserByUsername(ctx, creds.Username)
		if err != nil {
			t.Fatalf("GetUserByUsername: %v", err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(creds.TempPassword)); err != nil {
			t.Errorf("persisted hash does not match reissued temp password: %v", err)
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(reg.TempPassword)) == nil {
			t.Error("old temp password should no longer validate against the persisted hash")
		}
	})
}
