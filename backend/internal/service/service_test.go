package service

import (
	"context"
	"errors"
	"testing"

	"akademi-bimbel/internal/model"
)

// pingFailRepo overrides Ping to simulate a broken postgres connection.
type pingFailRepo struct {
	fakeUserRepo
}

func (f *pingFailRepo) Ping(_ context.Context) error {
	return errors.New("connection refused")
}

func TestHealth_AllUp(t *testing.T) {
	repo := newFakeUserRepo()
	svc, _ := newTestService(t, repo)

	h := svc.Health(context.Background())
	if h.Status != "ok" {
		t.Errorf("Status: want ok, got %s", h.Status)
	}
	if h.Postgres != "ok" {
		t.Errorf("Postgres: want ok, got %s", h.Postgres)
	}
	if h.Redis != "ok" {
		t.Errorf("Redis: want ok, got %s", h.Redis)
	}
}

func TestHealth_PostgresDown(t *testing.T) {
	svc, _ := newTestService(t, &pingFailRepo{})

	h := svc.Health(context.Background())
	if h.Status != "degraded" {
		t.Errorf("Status: want degraded, got %s", h.Status)
	}
	if h.Postgres != "down" {
		t.Errorf("Postgres: want down, got %s", h.Postgres)
	}
	if h.Redis != "ok" {
		t.Errorf("Redis: want ok, got %s", h.Redis)
	}
}

func TestHealth_RedisDown(t *testing.T) {
	repo := newFakeUserRepo()
	svc, mr := newTestService(t, repo)
	mr.Close() // force redis failures

	h := svc.Health(context.Background())
	if h.Status != "degraded" {
		t.Errorf("Status: want degraded, got %s", h.Status)
	}
	if h.Postgres != "ok" {
		t.Errorf("Postgres: want ok, got %s", h.Postgres)
	}
	if h.Redis != "down" {
		t.Errorf("Redis: want down, got %s", h.Redis)
	}
}

func TestParseAccess_ValidToken(t *testing.T) {
	repo := newFakeUserRepo()
	repo.seed(&model.User{
		ID:           "u-parse",
		Email:        strptr("parse@example.com"),
		PasswordHash: mustHashStd("password123"),
		Role:         RoleStudent,
		Status:       "active",
	})
	svc, _ := newTestService(t, repo)

	// Mint a token via the service's own login flow to get a valid JWT.
	access, _, err := svc.Login(context.Background(), "parse@example.com", "password123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if access == "" {
		t.Fatal("empty access token")
	}

	claims, err := svc.ParseAccess(access)
	if err != nil {
		t.Fatalf("ParseAccess: want nil err, got %v", err)
	}
	if claims.Sub != "u-parse" {
		t.Errorf("Sub: want u-parse, got %s", claims.Sub)
	}
}

func TestParseAccess_InvalidToken(t *testing.T) {
	repo := newFakeUserRepo()
	svc, _ := newTestService(t, repo)

	_, err := svc.ParseAccess("this.is.not.a.valid.token")
	if err == nil {
		t.Error("ParseAccess: want error for invalid token, got nil")
	}
}

func TestNewForTest(t *testing.T) {
	svc := NewForTest(nil)
	if svc == nil {
		t.Fatal("NewForTest: got nil")
	}
}

func TestNew(t *testing.T) {
	svc, _ := newTestService(t, newFakeUserRepo())
	if svc == nil {
		t.Fatal("New: got nil service")
	}
}
