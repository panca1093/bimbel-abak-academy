package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/service"
)

// registerExamResultRoutes adds the student result endpoint under /api/v1/exam,
// protected by JWTMiddleware.
func registerExamResultRoutes(t *testing.T, env *testEnv, h *handler.Handler) {
	t.Helper()
	v1 := env.e.Group("/api/v1")
	exam := v1.Group("/exam")
	exam.Use(handler.JWTMiddleware(env.svc, env.signer))
	exam.GET("/sessions/:id/result", h.StudentGetSessionResult)
}

// registerAdminGradingRoutes adds the admin grading endpoints (grading queue, essays
// read, grade write) under /api/v1/admin, protected by JWTMiddleware + RBACMiddleware.
func registerAdminGradingRoutes(t *testing.T, env *testEnv, h *handler.Handler) {
	t.Helper()
	v1 := env.e.Group("/api/v1")
	admin := v1.Group("/admin")
	admin.Use(handler.JWTMiddleware(env.svc, env.signer))

	adminExams := admin.Group("/exams")
	adminExams.Use(handler.RBACMiddleware("products(exam):write"))
	adminExams.GET("/:id/grading", h.AdminListGradingSessions)

	adminSessions := admin.Group("/sessions")
	adminSessions.Use(handler.RBACMiddleware("sessions:write"))
	adminSessions.GET("/:id/essays", h.AdminGetSessionEssays)
	adminSessions.POST("/:id/grade", h.AdminGradeEssay)
}

func postJSONWithToken(t *testing.T, e interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}, path, token string, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ---------- Student result endpoint ----------

func TestStudentGetSessionResult_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	registerExamResultRoutes(t, env, handler.New(env.svc))

	rec := getWithToken(t, env.e, "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000/result", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

// TestStudentGetSessionResult_InvalidStudentID_Returns422 proves the route resolves,
// JWTMiddleware accepts the token, and the handler forwards claims.Sub + the path param
// to the service — GetSessionResult validates studentID before touching storeRepo, so
// this exercises the full handler->service->mapServiceError chain without a real DB.
func TestStudentGetSessionResult_InvalidStudentID_Returns422(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "not-a-uuid-student",
		Email:  strptr("result-student@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerExamResultRoutes(t, env, h)

	token := mintToken(t, env, "not-a-uuid-student", service.RoleStudent)

	rec := getWithToken(t, env.e, "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000/result", token)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "validation_failed" {
		t.Errorf("code: want validation_failed, got %v", resp["code"])
	}
}

// ---------- Admin grading queue (list) ----------

func TestAdminListGradingSessions_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	rec := getWithToken(t, env.e, "/api/v1/admin/exams/00000000-0000-0000-0000-000000000000/grading", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminListGradingSessions_StudentToken_Returns403(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "student-grading-list",
		Email:  strptr("student-grading-list@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	token := mintToken(t, env, "student-grading-list", service.RoleStudent)

	rec := getWithToken(t, env.e, "/api/v1/admin/exams/00000000-0000-0000-0000-000000000000/grading", token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "forbidden" {
		t.Errorf("code: want forbidden, got %v", resp["code"])
	}
}

func TestAdminListGradingSessions_InvalidExamID_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "admin-grading-list",
		Email:  strptr("admin-grading-list@test.com"),
		Role:   service.RoleAdminExam,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	token := mintToken(t, env, "admin-grading-list", service.RoleAdminExam)

	rec := getWithToken(t, env.e, "/api/v1/admin/exams/not-a-uuid/grading", token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

// ---------- Admin session essays read ----------

func TestAdminGetSessionEssays_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	rec := getWithToken(t, env.e, "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/essays", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminGetSessionEssays_StudentToken_Returns403(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "student-essays",
		Email:  strptr("student-essays@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	token := mintToken(t, env, "student-essays", service.RoleStudent)

	rec := getWithToken(t, env.e, "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/essays", token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminGetSessionEssays_InvalidSessionID_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "admin-essays",
		Email:  strptr("admin-essays@test.com"),
		Role:   service.RoleAdminExam,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	token := mintToken(t, env, "admin-essays", service.RoleAdminExam)

	rec := getWithToken(t, env.e, "/api/v1/admin/sessions/not-a-uuid/essays", token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

// ---------- Admin grade essay write ----------

func TestAdminGradeEssay_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	rec := postJSONWithToken(t, env.e, "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/grade", "",
		`{"question_id":"00000000-0000-0000-0000-000000000000","score":1}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminGradeEssay_StudentToken_Returns403(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "student-grade",
		Email:  strptr("student-grade@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	token := mintToken(t, env, "student-grade", service.RoleStudent)

	rec := postJSONWithToken(t, env.e, "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/grade", token,
		`{"question_id":"00000000-0000-0000-0000-000000000000","score":1}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminGradeEssay_InvalidSessionID_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "admin-grade-badsession",
		Email:  strptr("admin-grade-badsession@test.com"),
		Role:   service.RoleAdminExam,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	token := mintToken(t, env, "admin-grade-badsession", service.RoleAdminExam)

	rec := postJSONWithToken(t, env.e, "/api/v1/admin/sessions/not-a-uuid/grade", token,
		`{"question_id":"00000000-0000-0000-0000-000000000000","score":1}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

func TestAdminGradeEssay_InvalidBody_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "admin-grade-badbody",
		Email:  strptr("admin-grade-badbody@test.com"),
		Role:   service.RoleAdminExam,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	token := mintToken(t, env, "admin-grade-badbody", service.RoleAdminExam)

	rec := postJSONWithToken(t, env.e, "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/grade", token,
		"not json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

func TestAdminGradeEssay_InvalidQuestionID_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "admin-grade-badquestion",
		Email:  strptr("admin-grade-badquestion@test.com"),
		Role:   service.RoleAdminExam,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminGradingRoutes(t, env, h)

	token := mintToken(t, env, "admin-grade-badquestion", service.RoleAdminExam)

	rec := postJSONWithToken(t, env.e, "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/grade", token,
		`{"question_id":"not-a-uuid","score":1}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}
