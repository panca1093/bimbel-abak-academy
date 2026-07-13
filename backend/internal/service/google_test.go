package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"akademi-bimbel/internal/model"
)

func TestGoogleLogin_CreateSetsGoogleProvider(t *testing.T) {
	// Fake Google tokeninfo endpoint.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(googleTokenInfo{
			Aud:           "google-client-id",
			Email:         "google-user@example.com",
			EmailVerified: "true",
			Name:          "Google User",
		})
	}))
	defer ts.Close()

	// Swap transport so calls to googleapis go to our fake server.
	orig := http.DefaultClient.Transport
	http.DefaultClient.Transport = &googleTokenInfoTransport{fakeURL: ts.URL}
	defer func() { http.DefaultClient.Transport = orig }()

	repo := newFakeUserRepo()
	svc, _ := newTestService(t, repo)

	access, _, err := svc.GoogleLogin(context.Background(), "fake-id-token")
	if err != nil {
		t.Fatalf("GoogleLogin: %v", err)
	}
	if access == "" {
		t.Fatal("empty access token")
	}

	u, _ := repo.GetUserByEmail(context.Background(), "google-user@example.com")
	if u == nil {
		t.Fatal("user not created")
	}
	if u.AuthProvider != "google" {
		t.Errorf("AuthProvider: want 'google', got '%s'", u.AuthProvider)
	}
}

func TestGoogleLogin_ExistingUserKeepsOriginalProvider(t *testing.T) {
	// Fake Google tokeninfo endpoint.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(googleTokenInfo{
			Aud:           "google-client-id",
			Email:         "existing@example.com",
			EmailVerified: "true",
			Name:          "Existing User",
		})
	}))
	defer ts.Close()

	orig := http.DefaultClient.Transport
	http.DefaultClient.Transport = &googleTokenInfoTransport{fakeURL: ts.URL}
	defer func() { http.DefaultClient.Transport = orig }()

	repo := newFakeUserRepo()
	repo.seed(&model.User{
		Email:        strptr("existing@example.com"),
		PasswordHash: mustHashStd("password123"),
		Role:         RoleStudent,
		Status:       "active",
		AuthProvider: "password",
	})
	svc, _ := newTestService(t, repo)

	_, _, err := svc.GoogleLogin(context.Background(), "fake-id-token")
	if err != nil {
		t.Fatalf("GoogleLogin: %v", err)
	}

	u, _ := repo.GetUserByEmail(context.Background(), "existing@example.com")
	if u == nil {
		t.Fatal("user not found")
	}
	if u.AuthProvider != "password" {
		t.Errorf("AuthProvider: want 'password' (unchanged), got '%s'", u.AuthProvider)
	}
}

// googleTokenInfoTransport rewrites all requests to googlesapis to a fake
// server while passing through every other request untouched.
type googleTokenInfoTransport struct {
	fakeURL string
}

func (t *googleTokenInfoTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Host, "googleapis.com") {
		newReq, _ := http.NewRequestWithContext(req.Context(), req.Method, t.fakeURL+"?id_token=fake", nil)
		return http.DefaultTransport.RoundTrip(newReq)
	}
	return http.DefaultTransport.RoundTrip(req)
}
