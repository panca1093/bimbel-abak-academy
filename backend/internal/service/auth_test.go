package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"akademi-bimbel/internal/config"
	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/repository"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct {
	byID    map[string]*repository.User
	seq     int
	createErr error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: map[string]*repository.User{}}
}

func (f *fakeUserRepo) Ping(_ context.Context) error { return nil }

func (f *fakeUserRepo) CreateUser(_ context.Context, u *repository.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.seq++
	u.ID = fmt.Sprintf("user-%d", f.seq)
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	cp := *u
	f.byID[u.ID] = &cp
	return nil
}

func (f *fakeUserRepo) GetUserByEmail(_ context.Context, email string) (*repository.User, error) {
	for _, u := range f.byID {
		if u.Email != nil && *u.Email == email && u.Status != "deleted" {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepo) GetUserByUsername(_ context.Context, username string) (*repository.User, error) {
	for _, u := range f.byID {
		if u.Username != nil && *u.Username == username && u.Status != "deleted" {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepo) GetUserByID(_ context.Context, id string) (*repository.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) UpdatePasswordHash(_ context.Context, userID, hash string) error {
	u, ok := f.byID[userID]
	if !ok {
		return errors.New("not found")
	}
	u.PasswordHash = hash
	return nil
}

func (f *fakeUserRepo) TombstoneUser(_ context.Context, userID string) error {
	u, ok := f.byID[userID]
	if !ok {
		return errors.New("not found")
	}
	u.Status = "deleted"
	return nil
}

// seed inserts a user directly, bypassing CreateUser sequencing concerns.
func (f *fakeUserRepo) seed(u *repository.User) {
	f.seq++
	if u.ID == "" {
		u.ID = fmt.Sprintf("seed-%d", f.seq)
	}
	cp := *u
	f.byID[u.ID] = &cp
}

func strptr(s string) *string { return &s }

func newTestService(t *testing.T, repo UserRepository) (*Service, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cfg := &config.Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 168 * time.Hour,
		OTPTTL:          5 * time.Minute,
		GoogleClientID:  "google-client-id",
	}
	signer := platform.NewJWTSigner(cfg.JWTSecret, cfg.AccessTokenTTL)
	svc := New(repo, rdb, signer, &platform.NoopOTPProvider{}, &platform.NoopEmailProvider{}, cfg)
	return svc, mr
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path creates user", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc, mr := newTestService(t, repo)
		pending, err := svc.Register(ctx, "New@Example.com", "password123", "Budi")
		if err != nil {
			t.Fatalf("Register: %v", err)
		}
		if pending == "" {
			t.Error("want non-empty pending_token")
		}
		u, _ := repo.GetUserByEmail(ctx, "new@example.com")
		if u == nil {
			t.Fatal("user not created")
		}
		if u.Role != RoleStudent || u.Status != "active" || !u.OTPEnabled {
			t.Errorf("unexpected user defaults: role=%s status=%s otp=%v", u.Role, u.Status, u.OTPEnabled)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("password123")); err != nil {
			t.Errorf("password not hashed correctly: %v", err)
		}
		if !mr.Exists("otp:" + u.ID) {
			t.Error("otp key should be stored after register")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&repository.User{Email: strptr("taken@example.com"), Status: "active", Role: RoleStudent})
		svc, _ := newTestService(t, repo)
		_, err := svc.Register(ctx, "taken@example.com", "password123", "Budi")
		if !errors.Is(err, ErrEmailTaken) {
			t.Errorf("want ErrEmailTaken, got %v", err)
		}
	})

	t.Run("weak password", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc, _ := newTestService(t, repo)
		_, err := svc.Register(ctx, "x@example.com", "short", "Budi")
		if !errors.Is(err, ErrWeakPassword) {
			t.Errorf("want ErrWeakPassword, got %v", err)
		}
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	seedActive := func(repo *fakeUserRepo, otp bool) {
		repo.seed(&repository.User{
			Email:        strptr("user@example.com"),
			PasswordHash: mustHashStd("password123"),
			Role:         RoleStudent,
			Status:       "active",
			OTPEnabled:   otp,
		})
	}

	t.Run("wrong password", func(t *testing.T) {
		repo := newFakeUserRepo()
		seedActive(repo, false)
		svc, _ := newTestService(t, repo)
		_, _, _, _, err := svc.Login(ctx, "user@example.com", "wrong")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("inactive user", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&repository.User{
			Email:        strptr("user@example.com"),
			PasswordHash: mustHashStd("password123"),
			Role:         RoleStudent,
			Status:       "deactivated",
		})
		svc, _ := newTestService(t, repo)
		_, _, _, _, err := svc.Login(ctx, "user@example.com", "password123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("otp disabled returns tokens inline", func(t *testing.T) {
		repo := newFakeUserRepo()
		seedActive(repo, false)
		svc, _ := newTestService(t, repo)
		pending, otpRequired, access, refresh, err := svc.Login(ctx, "user@example.com", "password123")
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		if otpRequired || pending != "" {
			t.Errorf("want otpRequired=false pending empty, got %v %q", otpRequired, pending)
		}
		if access == "" || refresh == "" {
			t.Error("want non-empty tokens")
		}
	})

	t.Run("otp enabled returns pending token", func(t *testing.T) {
		repo := newFakeUserRepo()
		seedActive(repo, true)
		svc, mr := newTestService(t, repo)
		pending, otpRequired, access, refresh, err := svc.Login(ctx, "user@example.com", "password123")
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		if !otpRequired || pending == "" {
			t.Errorf("want otpRequired=true and pending token, got %v %q", otpRequired, pending)
		}
		if access != "" || refresh != "" {
			t.Error("want empty tokens when otp required")
		}
		if got, _ := mr.Get("pending:" + pending); got == "" {
			t.Error("pending token not stored in redis")
		}
	})

	t.Run("login by username", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&repository.User{
			Username:     strptr("budi"),
			PasswordHash: mustHashStd("password123"),
			Role:         RoleStudent,
			Status:       "active",
		})
		svc, _ := newTestService(t, repo)
		_, _, access, _, err := svc.Login(ctx, "budi", "password123")
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		if access == "" {
			t.Error("want token for username login")
		}
	})
}

func TestVerifyOTP(t *testing.T) {
	ctx := context.Background()

	setup := func(t *testing.T) (*Service, *miniredis.Miniredis, *fakeUserRepo, string) {
		repo := newFakeUserRepo()
		repo.seed(&repository.User{
			ID:           "u1",
			Email:        strptr("user@example.com"),
			PasswordHash: mustHashStd("password123"),
			Role:         RoleStudent,
			Status:       "active",
			OTPEnabled:   true,
		})
		svc, mr := newTestService(t, repo)
		pending, _, _, _, err := svc.Login(ctx, "user@example.com", "password123")
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		return svc, mr, repo, pending
	}

	t.Run("correct code returns tokens and clears keys", func(t *testing.T) {
		svc, mr, _, pending := setup(t)
		code, _ := mr.Get("otp:u1")
		access, refresh, err := svc.VerifyOTP(ctx, pending, code)
		if err != nil {
			t.Fatalf("VerifyOTP: %v", err)
		}
		if access == "" || refresh == "" {
			t.Error("want tokens")
		}
		if mr.Exists("otp:u1") || mr.Exists("pending:"+pending) {
			t.Error("otp/pending keys should be deleted after verify")
		}
	})

	t.Run("wrong code", func(t *testing.T) {
		svc, _, _, pending := setup(t)
		_, _, err := svc.VerifyOTP(ctx, pending, "000000")
		if !errors.Is(err, ErrInvalidOTP) {
			t.Errorf("want ErrInvalidOTP, got %v", err)
		}
	})

	t.Run("expired otp", func(t *testing.T) {
		svc, mr, _, pending := setup(t)
		mr.Del("otp:u1")
		_, _, err := svc.VerifyOTP(ctx, pending, "123456")
		if !errors.Is(err, ErrOTPExpired) {
			t.Errorf("want ErrOTPExpired, got %v", err)
		}
	})

	t.Run("invalid pending token", func(t *testing.T) {
		svc, _, _, _ := setup(t)
		_, _, err := svc.VerifyOTP(ctx, "bogus", "123456")
		if !errors.Is(err, ErrInvalidPendingToken) {
			t.Errorf("want ErrInvalidPendingToken, got %v", err)
		}
	})
}

func TestRefresh(t *testing.T) {
	ctx := context.Background()

	t.Run("valid token rotates", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&repository.User{
			ID:           "u1",
			Email:        strptr("user@example.com"),
			PasswordHash: mustHashStd("password123"),
			Role:         RoleStudent,
			Status:       "active",
		})
		svc, mr := newTestService(t, repo)
		_, _, _, refresh, err := svc.Login(ctx, "user@example.com", "password123")
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		newAccess, newRefresh, err := svc.Refresh(ctx, refresh)
		if err != nil {
			t.Fatalf("Refresh: %v", err)
		}
		if newAccess == "" || newRefresh == "" {
			t.Error("want new tokens")
		}
		if mr.Exists("session:refresh:" + refresh) {
			t.Error("old refresh key should be deleted")
		}
		if !mr.Exists("session:refresh:" + newRefresh) {
			t.Error("new refresh key should exist")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc, _ := newTestService(t, repo)
		_, _, err := svc.Refresh(ctx, "nope")
		if !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("want ErrInvalidRefreshToken, got %v", err)
		}
	})
}

func TestLogout(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	repo.seed(&repository.User{
		ID:           "u1",
		Email:        strptr("user@example.com"),
		PasswordHash: mustHashStd("password123"),
		Role:         RoleStudent,
		Status:       "active",
	})
	svc, mr := newTestService(t, repo)
	_, _, access, _, err := svc.Login(ctx, "user@example.com", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	claims, err := platform.NewJWTSigner("test-secret", 15*time.Minute).ParseAccess(access)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	jti := claims.ID
	if !mr.Exists("session:access:" + jti) {
		t.Fatal("access session should exist")
	}
	if err := svc.Logout(ctx, jti, ""); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if mr.Exists("session:access:" + jti) {
		t.Error("access session should be deleted")
	}
	if err := svc.Logout(ctx, jti, ""); err != nil {
		t.Errorf("second Logout should be idempotent, got %v", err)
	}
}

func TestSendOTP(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	repo.seed(&repository.User{
		ID:         "u1",
		Email:      strptr("user@example.com"),
		Role:       RoleStudent,
		Status:     "active",
		OTPEnabled: true,
	})
	svc, mr := newTestService(t, repo)

	if err := svc.SendOTP(ctx, "u1"); err != nil {
		t.Fatalf("first SendOTP: %v", err)
	}
	if !mr.Exists("otp:u1") {
		t.Error("otp key should be set")
	}
	if err := svc.SendOTP(ctx, "u1"); !errors.Is(err, ErrOTPRateLimit) {
		t.Errorf("second SendOTP within window: want ErrOTPRateLimit, got %v", err)
	}
}

func TestResetPassword(t *testing.T) {
	ctx := context.Background()

	t.Run("valid token and otp updates password", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&repository.User{
			ID:           "u1",
			Email:        strptr("user@example.com"),
			PasswordHash: mustHashStd("oldpassword"),
			Role:         RoleStudent,
			Status:       "active",
		})
		svc, mr := newTestService(t, repo)
		mr.Set("reset:tok", "u1:654321")
		if err := svc.ResetPassword(ctx, "tok", "654321", "brandnewpass"); err != nil {
			t.Fatalf("ResetPassword: %v", err)
		}
		u, _ := repo.GetUserByID(ctx, "u1")
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("brandnewpass")); err != nil {
			t.Errorf("password not updated: %v", err)
		}
		if mr.Exists("reset:tok") {
			t.Error("reset token should be deleted")
		}
	})

	t.Run("wrong otp", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&repository.User{ID: "u1", Email: strptr("user@example.com"), PasswordHash: mustHashStd("oldpassword"), Role: RoleStudent, Status: "active"})
		svc, mr := newTestService(t, repo)
		mr.Set("reset:tok", "u1:654321")
		if err := svc.ResetPassword(ctx, "tok", "000000", "brandnewpass"); !errors.Is(err, ErrInvalidResetToken) {
			t.Errorf("want ErrInvalidResetToken, got %v", err)
		}
	})

	t.Run("invalid reset token", func(t *testing.T) {
		repo := newFakeUserRepo()
		svc, _ := newTestService(t, repo)
		if err := svc.ResetPassword(ctx, "missing", "654321", "brandnewpass"); !errors.Is(err, ErrInvalidResetToken) {
			t.Errorf("want ErrInvalidResetToken, got %v", err)
		}
	})
}

// TestGoogleLogin_InvalidToken covers FR-13: bad id_token → ErrInvalidToken.
func TestGoogleLogin_InvalidToken(t *testing.T) {
	// Stand up a fake tokeninfo server that returns 400.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	// Redirect http.DefaultClient to the fake server by overriding transport.
	orig := http.DefaultTransport
	http.DefaultTransport = &redirectTransport{target: ts.URL, transport: http.DefaultTransport}
	defer func() { http.DefaultTransport = orig }()

	repo := newFakeUserRepo()
	svc, _ := newTestService(t, repo)

	_, _, _, _, err := svc.GoogleLogin(context.Background(), "bogus-id-token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("want ErrInvalidToken, got %v", err)
	}
}

// redirectTransport rewrites all requests to a fixed base URL (fake server).
type redirectTransport struct {
	target    string
	transport http.RoundTripper
}

func (r *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := r.target + req.URL.RequestURI()
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, newURL, req.Body)
	if err != nil {
		return nil, err
	}
	return r.transport.RoundTrip(newReq)
}

// TestLogout_WithRefreshToken covers FR-17: logout also deletes the refresh token when supplied.
func TestLogout_WithRefreshToken(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	repo.seed(&repository.User{
		ID:           "u1",
		Email:        strptr("user@example.com"),
		PasswordHash: mustHashStd("password123"),
		Role:         RoleStudent,
		Status:       "active",
	})
	svc, mr := newTestService(t, repo)

	_, _, access, refresh, err := svc.Login(ctx, "user@example.com", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	claims, err := platform.NewJWTSigner("test-secret", 15*time.Minute).ParseAccess(access)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	jti := claims.ID

	if !mr.Exists("session:refresh:" + refresh) {
		t.Fatal("refresh key should exist before logout")
	}

	if err := svc.Logout(ctx, jti, refresh); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if mr.Exists("session:access:" + jti) {
		t.Error("access session should be deleted")
	}
	if mr.Exists("session:refresh:" + refresh) {
		t.Error("refresh session should be deleted when refresh_token supplied")
	}
}

// TestResetPassword_RevokesAllSessions covers FR-20: password reset invalidates ALL user sessions.
func TestResetPassword_RevokesAllSessions(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	repo.seed(&repository.User{
		ID:           "u1",
		Email:        strptr("user@example.com"),
		PasswordHash: mustHashStd("oldpassword"),
		Role:         RoleStudent,
		Status:       "active",
	})
	svc, mr := newTestService(t, repo)

	// Mint a session so we have live access+refresh keys.
	_, _, access, refresh, err := svc.Login(ctx, "user@example.com", "oldpassword")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	claims, err := platform.NewJWTSigner("test-secret", 15*time.Minute).ParseAccess(access)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	jti := claims.ID

	// Confirm keys exist before reset.
	if !mr.Exists("session:access:" + jti) {
		t.Fatal("access key should exist before reset")
	}
	if !mr.Exists("session:refresh:" + refresh) {
		t.Fatal("refresh key should exist before reset")
	}

	// Trigger reset.
	mr.Set("reset:tok", "u1:654321")
	if err := svc.ResetPassword(ctx, "tok", "654321", "brandnewpass"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	// Both session keys must be gone.
	if mr.Exists("session:access:" + jti) {
		t.Error("access session should be invalidated after password reset")
	}
	if mr.Exists("session:refresh:" + refresh) {
		t.Error("refresh session should be invalidated after password reset")
	}
}

// mustHashStd hashes outside a *testing.T context (used in closures/seed).
func mustHashStd(pw string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), 12)
	if err != nil {
		panic(err)
	}
	return string(h)
}
