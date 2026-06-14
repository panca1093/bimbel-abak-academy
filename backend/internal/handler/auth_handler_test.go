package handler_test

import (
	"akademi-bimbel/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/infra"
	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// fakeRepo is an in-memory UserRepository for handler tests.
type fakeRepo struct {
	byID map[string]*model.User
	seq  int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[string]*model.User{}}
}

func (f *fakeRepo) Ping(_ context.Context) error { return nil }

func (f *fakeRepo) CreateUser(_ context.Context, u *model.User) error {
	f.seq++
	u.ID = fmt.Sprintf("u%d", f.seq)
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	cp := *u
	f.byID[u.ID] = &cp
	return nil
}

func (f *fakeRepo) GetUserByEmail(_ context.Context, email string) (*model.User, error) {
	for _, u := range f.byID {
		if u.Email != nil && *u.Email == email && u.Status != "deleted" {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}

func (f *fakeRepo) GetUserByUsername(_ context.Context, username string) (*model.User, error) {
	for _, u := range f.byID {
		if u.Username != nil && *u.Username == username && u.Status != "deleted" {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}

func (f *fakeRepo) GetUserByID(_ context.Context, id string) (*model.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (f *fakeRepo) UpdatePasswordHash(_ context.Context, userID, hash string) error {
	u, ok := f.byID[userID]
	if !ok {
		return fmt.Errorf("not found")
	}
	u.PasswordHash = hash
	return nil
}

func (f *fakeRepo) TombstoneUser(_ context.Context, userID string) error {
	u, ok := f.byID[userID]
	if !ok {
		return fmt.Errorf("not found")
	}
	u.Status = "deleted"
	return nil
}

func (f *fakeRepo) seed(u *model.User) {
	f.seq++
	if u.ID == "" {
		u.ID = fmt.Sprintf("seed%d", f.seq)
	}
	cp := *u
	f.byID[u.ID] = &cp
}

func strptr(s string) *string { return &s }

func mustHash(pw string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), 12)
	if err != nil {
		panic(err)
	}
	return string(h)
}

type testEnv struct {
	e      *echo.Echo
	mr     *miniredis.Miniredis
	svc    *service.Service
	signer *infra.JWTSigner
	repo   *fakeRepo
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := &config.Config{
		JWTSecret:       "handler-test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 168 * time.Hour,
		OTPTTL:          5 * time.Minute,
	}
	signer := infra.NewJWTSigner(cfg.JWTSecret, cfg.AccessTokenTTL)
	repo := newFakeRepo()
	svc := service.New(repo, rdb, signer, &service.NoopOTPProvider{}, &service.NoopEmailProvider{}, cfg)

	h := handler.New(svc)
	e := echo.New()
	e.HideBanner = true
	v1 := e.Group("/api/v1")
	auth := v1.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login, handler.LoginRateLimiter())
	auth.POST("/otp/send", h.SendOTP)
	auth.POST("/otp/verify", h.VerifyOTP)
	auth.POST("/logout", h.Logout, handler.JWTMiddleware(svc, signer))
	auth.POST("/password/forgot", h.ForgotPassword)
	auth.POST("/password/reset", h.ResetPassword)
	auth.PATCH("/password/change", h.ChangePassword, handler.JWTMiddleware(svc, signer))
	auth.GET("/me", h.Me, handler.JWTMiddleware(svc, signer))

	admin := v1.Group("/admin")
	admin.Use(handler.JWTMiddleware(svc, signer))
	adminProducts := admin.Group("/products")
	adminProducts.GET("", h.AdminListProducts)
	adminProducts.POST("", h.AdminCreateProduct)
	adminProducts.GET("/:id", h.AdminGetProduct)
	adminProducts.PATCH("/:id", h.AdminUpdateProduct)
	adminProducts.POST("/:id/publish", h.AdminPublishProduct)
	adminProducts.DELETE("/:id", h.AdminDeleteProduct)

	return &testEnv{e: e, mr: mr, svc: svc, signer: signer, repo: repo}
}

func postJSON(t *testing.T, e *echo.Echo, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func getWithToken(t *testing.T, e *echo.Echo, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// --- Test cases ---

func TestRegisterHandler_HappyPath(t *testing.T) {
	env := newTestEnv(t)
	rec := postJSON(t, env.e, "/api/v1/auth/register", map[string]string{
		"email":    "budi@example.com",
		"password": "strongpass123",
		"name":     "Budi",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["otp_required"] != true {
		t.Errorf("want otp_required=true, got %v", resp["otp_required"])
	}
	if resp["pending_token"] == "" || resp["pending_token"] == nil {
		t.Error("want non-empty pending_token")
	}
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		Email:        strptr("taken@example.com"),
		PasswordHash: mustHash("whatever"),
		Role:         service.RoleStudent,
		Status:       "active",
	})
	rec := postJSON(t, env.e, "/api/v1/auth/register", map[string]string{
		"email":    "taken@example.com",
		"password": "strongpass123",
		"name":     "Budi",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "email_taken" {
		t.Errorf("want code=email_taken, got %v", resp["code"])
	}
}

func TestLoginHandler_OTPEnabled(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:           "u1",
		Email:        strptr("otp@example.com"),
		PasswordHash: mustHash("password123"),
		Role:         service.RoleStudent,
		Status:       "active",
		OTPEnabled:   true,
	})
	rec := postJSON(t, env.e, "/api/v1/auth/login", map[string]string{
		"identifier": "otp@example.com",
		"password":   "password123",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["otp_required"] != true {
		t.Errorf("want otp_required=true, got %v", resp["otp_required"])
	}
	if resp["pending_token"] == "" || resp["pending_token"] == nil {
		t.Error("want non-empty pending_token")
	}
}

func TestVerifyOTPHandler_HappyPath(t *testing.T) {
	env := newTestEnv(t)
	env.repo.seed(&model.User{
		ID:           "u1",
		Email:        strptr("otp@example.com"),
		PasswordHash: mustHash("password123"),
		Role:         service.RoleStudent,
		Status:       "active",
		OTPEnabled:   true,
	})

	// Login to get pending_token
	loginRec := postJSON(t, env.e, "/api/v1/auth/login", map[string]string{
		"identifier": "otp@example.com",
		"password":   "password123",
	})
	var loginResp map[string]any
	json.NewDecoder(loginRec.Body).Decode(&loginResp)
	pendingToken := loginResp["pending_token"].(string)

	// Get OTP from miniredis
	otpCode, _ := env.mr.Get("otp:u1")
	if otpCode == "" {
		t.Fatal("otp not stored in redis")
	}

	rec := postJSON(t, env.e, "/api/v1/auth/otp/verify", map[string]string{
		"pending_token": pendingToken,
		"code":          otpCode,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["access_token"] == "" || resp["access_token"] == nil {
		t.Error("want access_token")
	}
	if resp["refresh_token"] == "" || resp["refresh_token"] == nil {
		t.Error("want refresh_token")
	}
}

func TestMeHandler_WithoutToken(t *testing.T) {
	env := newTestEnv(t)
	rec := getWithToken(t, env.e, "/api/v1/auth/me", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "unauthorized" {
		t.Errorf("want code=unauthorized, got %v", resp["code"])
	}
}

// TestRegisterHandler_WeakPassword covers FR-3: weak password → 400 invalid_request.
func TestRegisterHandler_WeakPassword(t *testing.T) {
	env := newTestEnv(t)
	rec := postJSON(t, env.e, "/api/v1/auth/register", map[string]string{
		"email":    "budi@example.com",
		"password": "short",
		"name":     "Budi",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["code"] != "invalid_request" {
		t.Errorf("want code=invalid_request, got %v", resp["code"])
	}
}

// TestLoginHandler_RateLimit covers FR-29: >10 login attempts/min/IP → 429 rate_limited.
func TestLoginHandler_RateLimit(t *testing.T) {
	env := newTestEnv(t)
	body := map[string]string{"identifier": "x@example.com", "password": "wrongpass"}
	var lastCode int
	for i := 0; i < 12; i++ {
		rec := postJSON(t, env.e, "/api/v1/auth/login", body)
		lastCode = rec.Code
		if rec.Code == http.StatusTooManyRequests {
			var resp map[string]any
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp["code"] != "rate_limited" {
				t.Errorf("want code=rate_limited, got %v", resp["code"])
			}
			return
		}
	}
	t.Errorf("want 429 after 11th attempt, last code=%d", lastCode)
}

func TestMeHandler_ValidToken(t *testing.T) {
	env := newTestEnv(t)
	email := "me@example.com"
	env.repo.seed(&model.User{
		ID:           "u1",
		Email:        &email,
		PasswordHash: mustHash("password123"),
		Role:         service.RoleStudent,
		Status:       "active",
	})

	// Login to get access token
	loginRec := postJSON(t, env.e, "/api/v1/auth/login", map[string]string{
		"identifier": email,
		"password":   "password123",
	})
	var loginResp map[string]any
	json.NewDecoder(loginRec.Body).Decode(&loginResp)
	accessToken, ok := loginResp["access_token"].(string)
	if !ok || accessToken == "" {
		t.Fatalf("want access_token in login response, got %v", loginResp)
	}

	rec := getWithToken(t, env.e, "/api/v1/auth/me", accessToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["email"] != email {
		t.Errorf("want email=%s, got %v", email, resp["email"])
	}
	if resp["role"] == nil {
		t.Error("want role in response")
	}
}
