package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"akademi-bimbel/internal/service"
)

// ---------------------------------------------------------------------------
// AdminListExamRegistrations tests (FR-32 admin participant roster)
// ---------------------------------------------------------------------------

func TestAdminListExamRegistrations_NoToken_Returns401(t *testing.T) {
	env := newTestEnvWithStore(t)
	rec := getRequest(t, env.e, "/api/v1/admin/exams/00000000-0000-0000-0000-000000000000/registrations", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// admin_store has no products(exam) capability at all (see rbac.go) — the
// read-only roster endpoint must reject it same as the write-gated group.
func TestAdminListExamRegistrations_RoleWithoutReadCapability_Returns403(t *testing.T) {
	env := newTestEnvWithStore(t)
	admin := seedUser(t, env.pool, service.RoleAdminStore, "Store Admin")
	token := mintTokenForEnv(t, env, admin.String(), service.RoleAdminStore)

	examID := seedExam(t, env.pool, "Roster RBAC Exam", false, "hidden", "classic")

	rec := getRequest(t, env.e, "/api/v1/admin/exams/"+examID.String()+"/registrations", token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// admin_school only has products(exam):read (not :write) and must still be
// able to use the roster on the Registrations tab.
func TestAdminListExamRegistrations_AdminSchoolToken_Returns200WithRoster(t *testing.T) {
	env := newTestEnvWithStore(t)
	admin := seedUser(t, env.pool, service.RoleAdminSchool, "School Admin")
	token := mintTokenForEnv(t, env, admin.String(), service.RoleAdminSchool)

	examID := seedExam(t, env.pool, "Roster Exam", false, "hidden", "classic")
	student := seedUser(t, env.pool, service.RoleStudent, "Student Roster")
	regID := seedRegistration(t, env.pool, student, examID)
	if _, err := env.pool.Exec(context.Background(),
		`UPDATE exam_registration SET participant_number = 1 WHERE id = $1`, regID,
	); err != nil {
		t.Fatalf("set participant_number: %v", err)
	}

	rec := getRequest(t, env.e, "/api/v1/admin/exams/"+examID.String()+"/registrations", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("data is not an array: %T", resp["data"])
	}
	if len(data) != 1 {
		t.Fatalf("want 1 roster row, got %d", len(data))
	}
	row := data[0].(map[string]any)
	if row["student_id"] != student.String() {
		t.Errorf("student_id: want %s, got %v", student.String(), row["student_id"])
	}
	no, _ := row["participant_no"].(string)
	if no == "" {
		t.Errorf("want a non-empty participant_no for a row with participant_number set, got %v", row["participant_no"])
	}
}
