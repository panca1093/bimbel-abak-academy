package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/service"
)

// mintToken creates a signed JWT and stores the session in the miniredis-backed
// Redis client so the JWTMiddleware will accept it.
func mintToken(t *testing.T, env *testEnv, userID, role string) string {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: env.mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	tokenString, jti, err := env.signer.SignAccess(userID, role, nil, []string{})
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}
	if err := rdb.Set(context.Background(), "session:access:"+jti, userID, 15*time.Minute).Err(); err != nil {
		t.Fatalf("redis set session: %v", err)
	}
	return tokenString
}

// registerStudentRoutes adds exam session endpoints under /api/v1/exam on the
// given Echo instance, protected by JWTMiddleware.
func registerStudentRoutes(t *testing.T, env *testEnv, h *handler.Handler) {
	t.Helper()
	v1 := env.e.Group("/api/v1")
	exam := v1.Group("/exam")
	exam.Use(handler.JWTMiddleware(env.svc, env.signer))
	exam.POST("/checkin", h.StudentCheckIn)
	exam.POST("/sessions", h.StudentStartSession)
	exam.GET("/sessions/:id", h.StudentReconnectSession)
	exam.PATCH("/sessions/:id/answers", h.StudentSaveAnswers)
	exam.POST("/sessions/:id/submit", h.StudentSubmitSession)
	exam.POST("/sessions/:id/violations", h.StudentLogViolation)
}

// registerAdminSessionRoutes adds admin session endpoints under /api/v1/admin
// protected by JWTMiddleware + RBACMiddleware("sessions:write").
func registerAdminSessionRoutes(t *testing.T, env *testEnv, h *handler.Handler) {
	t.Helper()
	v1 := env.e.Group("/api/v1")
	admin := v1.Group("/admin")
	admin.Use(handler.JWTMiddleware(env.svc, env.signer))
	adminSessions := admin.Group("/sessions")
	adminSessions.Use(handler.RBACMiddleware("sessions:write"))
	adminSessions.POST("/:id/reopen", h.AdminReopenSession)
	adminSessions.POST("/:id/force-submit", h.AdminForceSubmitSession)
}

// ---------- Check-in tests ----------

func TestCheckIn_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	registerStudentRoutes(t, env, handler.New(env.svc))

	rec := postJSON(t, env.e, "/api/v1/exam/checkin", map[string]string{"token": "ABC123"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

func TestCheckIn_InvalidBody_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "checkin-user",
		Email:  strptr("checkin@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerStudentRoutes(t, env, h)

	token := mintToken(t, env, "checkin-user", service.RoleStudent)

	req := httptest.NewRequest("POST", "/api/v1/exam/checkin", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

// ---------- Session start tests ----------

func TestExamSessionStart_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	registerStudentRoutes(t, env, handler.New(env.svc))

	rec := postJSON(t, env.e, "/api/v1/exam/sessions", map[string]string{"registration_id": "00000000-0000-0000-0000-000000000000"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

func TestExamSessionStart_InvalidBody_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "start-user",
		Email:  strptr("start@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerStudentRoutes(t, env, h)

	token := mintToken(t, env, "start-user", service.RoleStudent)

	req := httptest.NewRequest("POST", "/api/v1/exam/sessions", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

// ---------- Reconnect tests ----------

func TestExamReconnect_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	registerStudentRoutes(t, env, handler.New(env.svc))

	req := httptest.NewRequest("GET", "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

// ---------- Save answers tests ----------

func TestExamSaveAnswers_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	registerStudentRoutes(t, env, handler.New(env.svc))

	rec := postJSON(t, env.e, "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000/answers",
		map[string]any{"answers": []map[string]any{}})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

func TestExamSaveAnswers_InvalidBody_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "save-user",
		Email:  strptr("save@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerStudentRoutes(t, env, h)

	token := mintToken(t, env, "save-user", service.RoleStudent)

	req := httptest.NewRequest("PATCH", "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000/answers",
		strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

// ---------- Submit tests ----------

func TestExamSubmit_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	registerStudentRoutes(t, env, handler.New(env.svc))

	req := httptest.NewRequest("POST", "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000/submit", nil)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

// ---------- Violation tests ----------

func TestExamViolation_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	registerStudentRoutes(t, env, handler.New(env.svc))

	rec := postJSON(t, env.e, "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000/violations",
		map[string]string{"violation_type": "tab_switch"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

func TestExamViolation_InvalidBody_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "violation-user",
		Email:  strptr("violation@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerStudentRoutes(t, env, h)

	token := mintToken(t, env, "violation-user", service.RoleStudent)

	req := httptest.NewRequest("POST", "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000/violations",
		strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

// TestExamViolation_InvalidType_Returns400 verifies that an unknown
// violation_type produces HTTP 400 with code "invalid_violation_type".
// The service checks validViolationTypes before accessing storeRepo, so this
// test works without a real database.
func TestExamViolation_InvalidType_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "violation-user-2",
		Email:  strptr("violation2@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerStudentRoutes(t, env, h)

	token := mintToken(t, env, "violation-user-2", service.RoleStudent)

	body := `{"violation_type":"unknown_type"}`
	req := httptest.NewRequest("POST", "/api/v1/exam/sessions/00000000-0000-0000-0000-000000000000/violations",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_violation_type" {
		t.Errorf("code: want invalid_violation_type, got %v", resp["code"])
	}
}

// ---------- Admin reopen tests ----------

func TestAdminSessionReopen_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	h := handler.New(env.svc)
	registerAdminSessionRoutes(t, env, h)

	rec := postJSON(t, env.e, "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/reopen",
		map[string]int{"extend_minutes": 30})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

func TestAdminSessionReopen_StudentToken_Returns403(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "student-reopen",
		Email:  strptr("student-reopen@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminSessionRoutes(t, env, h)

	token := mintToken(t, env, "student-reopen", service.RoleStudent)

	body := `{"extend_minutes": 30}`
	req := httptest.NewRequest("POST", "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/reopen",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "forbidden" {
		t.Errorf("code: want forbidden, got %v", resp["code"])
	}
}

// TestAdminSessionReopen_InvalidBody_Returns400 verifies that an admin_exam
// token passes both JWTMiddleware and RBACMiddleware, then the handler's c.Bind
// fails before calling the service — no storeRepo needed.
func TestAdminSessionReopen_InvalidBody_Returns400(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "admin-reopen",
		Email:  strptr("admin-reopen@test.com"),
		Role:   service.RoleAdminExam,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminSessionRoutes(t, env, h)

	token := mintToken(t, env, "admin-reopen", service.RoleAdminExam)

	req := httptest.NewRequest("POST", "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/reopen",
		strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

// ---------- Admin force-submit tests ----------

func TestAdminSessionForceSubmit_NoToken_Returns401(t *testing.T) {
	env := newTestEnv(t)
	h := handler.New(env.svc)
	registerAdminSessionRoutes(t, env, h)

	req := httptest.NewRequest("POST", "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/force-submit", nil)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("code: want unauthorized, got %v", resp["code"])
	}
}

func TestAdminSessionForceSubmit_StudentToken_Returns403(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:     "student-force",
		Email:  strptr("student-force@test.com"),
		Role:   service.RoleStudent,
		Status: "active",
	})
	h := handler.New(env.svc)
	registerAdminSessionRoutes(t, env, h)

	token := mintToken(t, env, "student-force", service.RoleStudent)

	req := httptest.NewRequest("POST", "/api/v1/admin/sessions/00000000-0000-0000-0000-000000000000/force-submit", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "forbidden" {
		t.Errorf("code: want forbidden, got %v", resp["code"])
	}
}
