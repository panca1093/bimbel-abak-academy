package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"akademi-bimbel/config"
	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/infra"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// setAdminSchoolClaims sets admin_school claims with a schoolID on the echo context.
func setAdminSchoolClaims(c echo.Context, sub, schoolID string) {
	c.Set("claims", &infra.Claims{Sub: sub, Role: "admin_school", SchoolID: &schoolID})
}

// setAdminSchoolClaimsNil sets admin_school claims with nil schoolID.
func setAdminSchoolClaimsNil(c echo.Context, sub string) {
	c.Set("claims", &infra.Claims{Sub: sub, Role: "admin_school", SchoolID: nil})
}

func TestAdminListStudents_NilSchoolID_403(t *testing.T) {
	env := newAdminSystemEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/students", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaimsNil(c, "u1")

	err := env.h.AdminListStudents(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for nil schoolID, got %d", rec.Code)
	}
}

func TestAdminRegisterStudent_NilSchoolID_403(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]string{"name": "Test", "nis": "12345"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaimsNil(c, "u1")

	err := env.h.AdminRegisterStudent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for nil schoolID, got %d", rec.Code)
	}
}

func TestAdminRegisterStudent_MissingFields(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]string{"nis": "12345"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaims(c, "u1", "s1")

	err := env.h.AdminRegisterStudent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 for missing name, got %d", rec.Code)
	}
}

func TestAdminRegisterStudent_MissingNIS(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]string{"name": "Test Student"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaims(c, "u1", "s1")

	err := env.h.AdminRegisterStudent(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 for missing nis, got %d", rec.Code)
	}
}

func TestAdminChangeStudentStatus_InvalidStatus(t *testing.T) {
	env := newAdminSystemEnv(t)

	body := map[string]string{"status": "deleted"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/admin/students/stu1", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("stu1")
	setAdminSchoolClaims(c, "u1", "s1")

	err := env.h.AdminChangeStudentStatus(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 for invalid status, got %d", rec.Code)
	}
}

func TestAdminGetStudentCredentials_NilSchoolID_403(t *testing.T) {
	env := newAdminSystemEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/students/stu1/credentials", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("stu1")
	setAdminSchoolClaimsNil(c, "u1")

	err := env.h.AdminGetStudentCredentials(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for nil schoolID, got %d", rec.Code)
	}
}

func TestAdminListStudents_SuperAdmin_MissingSchoolID_400(t *testing.T) {
	env := newAdminSystemEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/students", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminClaims(c, "u1") // super_admin, no school_id param

	if err := env.h.AdminListStudents(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 for super_admin without school_id, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("code: want invalid_request, got %v", resp["code"])
	}
}

func TestAdminListStudents_AdminSchool_DifferentSchoolID_403(t *testing.T) {
	env := newAdminSystemEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/students?school_id=s2", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaims(c, "u1", "s1") // admin_school with JWT school_id=s1, but query says s2

	if err := env.h.AdminListStudents(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403 for admin_school passing different school_id, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "forbidden" {
		t.Errorf("code: want forbidden, got %v", resp["code"])
	}
	if resp["message"] != "cannot widen school scope" {
		t.Errorf("message: want 'cannot widen school scope', got %v", resp["message"])
	}
}

func TestAdminListStudents_AdminSchool_SameSchoolID_Ignored(t *testing.T) {
	env := newAdminSystemEnv(t)

	// admin_school should ignore a school_id param matching their own scope.
	// This test uses RegisterStudent (enters the resolver path, then validates
	// request body — missing name means 400, not 403).
	body := map[string]string{"nis": "12345"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/admin/students?school_id=s1", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminSchoolClaims(c, "u1", "s1")

	if err := env.h.AdminRegisterStudent(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code == http.StatusForbidden {
		t.Errorf("admin_school with matching school_id should NOT get 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	// Should be 400 because name is missing, not 403 from scope rejection.
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400 for missing name, got %d", rec.Code)
	}
}

func TestAdminListStudents_RBAC_403(t *testing.T) {
	env := newAdminSystemEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/students", nil)
	rec := httptest.NewRecorder()
	c := env.e.NewContext(req, rec)
	setAdminClaims(c, "u1") // super_admin without schoolID — should be rejected

	// Route guard: RBACMiddleware("students:*") allows admin_school but
	// AdminListStudents checks Claims.SchoolID != nil.
	// Without a proper token with admin_school + schoolID, we get 403.
	err := env.h.AdminListStudents(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code == http.StatusOK {
		t.Error("admin_school without schoolID should be rejected")
	}
}

// ---------------------------------------------------------------------------
// DB-backed tests for school scope (students)
// ---------------------------------------------------------------------------

var (
	adminStuDBOnce sync.Once
	adminStuDBEnv  *adminStuDBTestEnv
)

type adminStuDBTestEnv struct {
	pool   *pgxpool.Pool
	e      *echo.Echo
	svc    *service.Service
	signer *infra.JWTSigner
	mr     *miniredis.Miniredis
}

func newAdminStuDBEnv(t *testing.T) *adminStuDBTestEnv {
	t.Helper()
	adminStuDBOnce.Do(func() {
		ctx := context.Background()

		pgContainer, err := tcpostgres.Run(ctx,
			"postgres:16-alpine",
			tcpostgres.WithDatabase("akademi_admin_stu_test"),
			tcpostgres.WithUsername("test"),
			tcpostgres.WithPassword("test"),
			tcpostgres.BasicWaitStrategies(),
		)
		if err != nil {
			t.Fatalf("start postgres container: %v", err)
		}
		dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			t.Fatalf("connection string: %v", err)
		}
		if err := infra.RunMigrations(ctx, dsn); err != nil {
			t.Fatalf("run migrations: %v", err)
		}
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			t.Fatalf("new pool: %v", err)
		}
		store := repository.New(pool)

		mr, err := miniredis.Run()
		if err != nil {
			t.Fatalf("miniredis: %v", err)
		}
		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

		cfg := &config.Config{
			JWTSecret:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 168 * time.Hour,
			OTPTTL:          5 * time.Minute,
		}
		signer := infra.NewJWTSigner(cfg.JWTSecret, cfg.AccessTokenTTL)
		svc := service.NewWithStore(
			store, store, rdb, signer,
			&service.NoopOTPProvider{}, &service.NoopEmailProvider{},
			&service.NoopPaymentClient{}, &service.NoopLogisticsClient{},
			nil, cfg,
		)
		e := echo.New()
		e.HideBanner = true
		h := handler.New(svc)

		// Register admin students routes.
		v1 := e.Group("/api/v1")
		admin := v1.Group("/admin")
		admin.Use(handler.JWTMiddleware(svc, signer))
		adminStudents := admin.Group("/students")
		adminStudents.Use(handler.RBACMiddleware("students:*"))
		adminStudents.GET("", h.AdminListStudents)
		adminStudents.POST("", h.AdminRegisterStudent)
		adminStudents.PATCH("/:id", h.AdminChangeStudentStatus)
		adminStudents.GET("/:id/credentials", h.AdminGetStudentCredentials)
		adminStudents.POST("/bulk/presign", h.AdminPresignStudentBulkUpload)
		adminStudents.POST("/bulk", h.AdminBulkImportStudents)
		adminStudents.POST("/bulk/credentials", h.AdminBulkReissueCredentials)

		adminStuDBEnv = &adminStuDBTestEnv{
			pool:   pool,
			e:      e,
			svc:    svc,
			signer: signer,
			mr:     mr,
		}
	})
	if adminStuDBEnv == nil {
		t.Fatal("admin students test env failed to initialize")
	}
	return adminStuDBEnv
}

func seedSchoolForStu(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := pool.QueryRow(ctx,
		`INSERT INTO school (name, code) VALUES ($1, $2) RETURNING id`,
		"Stu School", "stu_"+time.Now().Format("150405"),
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert school: %v", err)
	}
	return id
}

func mintSuperAdminStuToken(t *testing.T, env *adminStuDBTestEnv, userID string) string {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: env.mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	tokenString, jti, err := env.signer.SignAccess(userID, "super_admin", nil, []string{})
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}
	if err := rdb.Set(context.Background(), "session:access:"+jti, userID, 15*time.Minute).Err(); err != nil {
		t.Fatalf("redis set session: %v", err)
	}
	return tokenString
}

func TestAdminListStudents_SuperAdmin_NonexistentSchool_404(t *testing.T) {
	env := newAdminStuDBEnv(t)

	superToken := mintSuperAdminStuToken(t, env, "super1")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/students?school_id="+"00000000-0000-0000-0000-000000000000", nil)
	req.Header.Set("Authorization", "Bearer "+superToken)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 for nonexistent school, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "not_found" {
		t.Errorf("code: want not_found, got %v", resp["code"])
	}
}

func TestAdminListStudents_SuperAdmin_ValidSchool_200(t *testing.T) {
	env := newAdminStuDBEnv(t)

	schoolID := seedSchoolForStu(t, env.pool)
	superToken := mintSuperAdminStuToken(t, env, "super2")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/students?school_id="+schoolID, nil)
	req.Header.Set("Authorization", "Bearer "+superToken)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["data"] == nil {
		t.Error("want non-nil data array")
	}
}

func TestAdminListStudents_AdminSchool_OwnScope_200(t *testing.T) {
	env := newAdminStuDBEnv(t)

	schoolID := seedSchoolForStu(t, env.pool)

	// Mint an admin_school token scoped to the school.
	rdb := redis.NewClient(&redis.Options{Addr: env.mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	tokenString, jti, err := env.signer.SignAccess("admin1", "admin_school", &schoolID, []string{})
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}
	if err := rdb.Set(context.Background(), "session:access:"+jti, "admin1", 15*time.Minute).Err(); err != nil {
		t.Fatalf("redis set session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/students", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	env.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 for admin_school with own scope, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["data"] == nil {
		t.Error("want non-nil data array")
	}
}
