package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/service"
)

// registerStudentCardRoute adds the student card-download endpoint under
// /api/v1/exam, protected by JWTMiddleware only (no RBAC) — mirrors
// registerStudentLeaderboardRoute.
func registerStudentCardRoute(t *testing.T, env *testEnv, h *handler.Handler) {
	t.Helper()
	v1 := env.e.Group("/api/v1")
	exam := v1.Group("/exam")
	exam.Use(handler.JWTMiddleware(env.svc, env.signer))
	exam.GET("/registrations/:id/card", h.StudentGetExamCard)
}

// ---------------------------------------------------------------------------
// StudentGetExamCard tests (FR-30)
// ---------------------------------------------------------------------------

func TestStudentGetExamCard_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	h := handler.New(env.svc)
	registerStudentCardRoute(t, env, h)

	rec := getRequest(t, env.e, "/api/v1/exam/registrations/00000000-0000-0000-0000-000000000000/card", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestStudentGetExamCard_NotOwned_Returns404(t *testing.T) {
	env := newTestEnvWithStore(t)
	owner := seedUser(t, env.pool, "student", "Card Owner")
	other := seedUser(t, env.pool, "student", "Other Card Student")

	examID := seedExam(t, env.pool, "Card Exam", false, "hidden", "classic")
	regID := seedRegistration(t, env.pool, owner, examID)

	token := mintTokenForEnv(t, env, other.String(), service.RoleStudent)
	rec := getRequest(t, env.e, "/api/v1/exam/registrations/"+regID.String()+"/card", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "registration_not_found" {
		t.Errorf("code: want registration_not_found, got %v", resp["code"])
	}
}

// NOTE: no success/200 test here. Unlike certificate preview (a pure render,
// no persistence), GetExamCard's success path unconditionally reads/writes
// through *Service.storage (a concrete *minio.Client field, not an
// interface) via uploadCardPDF/downloadCardPDF — it cannot be faked without
// either a production interface seam or a live/testcontainer S3 backend,
// neither of which any existing test in this repo uses. See task concerns.
