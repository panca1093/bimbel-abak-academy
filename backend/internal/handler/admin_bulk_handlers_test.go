package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Minimal guard-clause tests driving Task 7's implementation. The definitive
// handler test files (admin_students_bulk_test.go / admin_jobs_test.go) are
// written in a later task with real storeRepo fixtures.

func TestAdminPresignStudentBulkUpload_NilSchoolID_403(t *testing.T) {
	env := newAdminSystemEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/admin/students/bulk/presign?filename=x.csv", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaimsNil(c, "u1")

	if err := env.h.AdminPresignStudentBulkUpload(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for nil schoolID, got %d", rec.Code)
	}
}

func TestAdminPresignStudentBulkUpload_MissingFilename_400(t *testing.T) {
	env := newAdminSystemEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/admin/students/bulk/presign", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaims(c, "u1", "s1")

	if err := env.h.AdminPresignStudentBulkUpload(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 for missing filename, got %d", rec.Code)
	}
}

func TestAdminBulkImportStudents_NilSchoolID_403(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]string{"file_key": "student-bulk/s1/x.csv"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students/bulk", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaimsNil(c, "u1")

	if err := env.h.AdminBulkImportStudents(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for nil schoolID, got %d", rec.Code)
	}
}

func TestAdminBulkImportStudents_MissingFileKey_400(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]string{}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students/bulk", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaims(c, "u1", "s1")

	if err := env.h.AdminBulkImportStudents(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 for missing file_key, got %d", rec.Code)
	}
}

func TestAdminBulkReissueCredentials_NilSchoolID_403(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]any{"all": true}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students/bulk/credentials", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaimsNil(c, "u1")

	if err := env.h.AdminBulkReissueCredentials(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for nil schoolID, got %d", rec.Code)
	}
}

func TestAdminBulkReissueCredentials_BothSupplied_400(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]any{"all": true, "student_ids": []string{"stu1"}}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students/bulk/credentials", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaims(c, "u1", "s1")

	if err := env.h.AdminBulkReissueCredentials(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 when both student_ids and all supplied, got %d", rec.Code)
	}
}

func TestAdminBulkReissueCredentials_NeitherSupplied_400(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]any{}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students/bulk/credentials", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaims(c, "u1", "s1")

	if err := env.h.AdminBulkReissueCredentials(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 when neither student_ids nor all supplied, got %d", rec.Code)
	}
}
