package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/service"

	"github.com/redis/go-redis/v9"
)

// TestAdminGrantExamAccess_RBAC verifies that non-super_admin roles get 403
// (FR-GRANT-05) and unauthenticated requests get 401.
func TestAdminGrantExamAccess_RBAC(t *testing.T) {
	env := newTestEnv(t)
	h := handler.New(env.svc)

	// Register the exam-grant route with RBAC middleware
	// (mirrors the pattern from server/routes.go's admin routes).
	admin := env.e.Group("/api/v1/admin")
	admin.Use(handler.JWTMiddleware(env.svc, env.signer))
	adminExamGrants := admin.Group("/exam-grants")
	adminExamGrants.Use(handler.RBACMiddleware("exam-grants:write"))
	adminExamGrants.POST("", h.AdminGrantExamAccess)

	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/exam-grants",
			strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		env.e.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("non-super_admin role gets 403", func(t *testing.T) {
		token := mintAccessToken(t, env, "admin-store-1", service.RoleAdminStore, nil)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/exam-grants",
			strings.NewReader(`{"exam_id":"00000000-0000-0000-0000-000000000001","student_ids":["00000000-0000-0000-0000-000000000002"]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		env.e.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
		}
		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp["code"] != "forbidden" {
			t.Errorf("code: want forbidden, got %s", resp["code"])
		}
	})
}

// mintAccessToken creates a JWT for testing, stored in Redis (required by JWTMiddleware).
func mintAccessToken(t *testing.T, env *testEnv, sub, role string, schoolID *string) string {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: env.mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	tokenString, jti, err := env.signer.SignAccess(sub, role, schoolID, []string{})
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}
	if err := rdb.Set(context.Background(), "session:access:"+jti, sub, 15*time.Minute).Err(); err != nil {
		t.Fatalf("redis set session: %v", err)
	}
	return tokenString
}
