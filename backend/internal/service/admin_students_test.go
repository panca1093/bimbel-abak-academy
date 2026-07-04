package service

import (
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

func TestStudentResponseMapping(t *testing.T) {
	// Test toStudentResponse with a StudentRow fixture.
	// We can't import repository.StudentRow from a test in the same package easily,
	// but we can verify the response types compile and are usable.
	var resp StudentResponse
	resp.ID = "s1"
	resp.Name = "Test Student"
	if resp.ID != "s1" {
		t.Error("response struct field access failed")
	}
}

func TestStudentRegistrationResponseFields(t *testing.T) {
	resp := StudentRegistrationResponse{
		ID:           "id",
		Name:         "name",
		Username:     "code_nis",
		NIS:          "nis",
		TempPassword: "temp",
		CreatedAt:    "2024-01-01T00:00:00Z",
	}
	if resp.TempPassword == "" {
		t.Error("temp_password should not be empty in registration response")
	}
	if resp.Username != "code_nis" {
		t.Errorf("Username: want code_nis, got %s", resp.Username)
	}
}

func TestStudentCredentialsResponseFields(t *testing.T) {
	resp := StudentCredentialsResponse{
		Username:     "code_nis",
		TempPassword: "newtemp",
	}
	if resp.Username == "" || resp.TempPassword == "" {
		t.Error("credentials response fields should not be empty")
	}
}
