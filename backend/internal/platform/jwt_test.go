package platform_test

import (
	"strings"
	"testing"
	"time"

	"akademi-bimbel/internal/platform"
)

func TestJWT_RoundTrip(t *testing.T) {
	signer := platform.NewJWTSigner("supersecret", 15*time.Minute)
	schoolID := "school-1"
	caps := []string{"read", "write"}

	token, _, err := signer.SignAccess("user-42", "admin", &schoolID, caps)
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}

	claims, err := signer.ParseAccess(token)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}

	if claims.Sub != "user-42" {
		t.Errorf("Sub: got %q want %q", claims.Sub, "user-42")
	}
	if claims.Role != "admin" {
		t.Errorf("Role: got %q want %q", claims.Role, "admin")
	}
	if claims.SchoolID == nil || *claims.SchoolID != "school-1" {
		t.Errorf("SchoolID: got %v want school-1", claims.SchoolID)
	}
	if len(claims.Capabilities) != 2 || claims.Capabilities[0] != "read" {
		t.Errorf("Capabilities: got %v", claims.Capabilities)
	}
}

func TestJWT_TamperedToken(t *testing.T) {
	signer := platform.NewJWTSigner("supersecret", 15*time.Minute)
	token, _, err := signer.SignAccess("user-1", "student", nil, nil)
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}

	// flip the last character of the signature
	tampered := token[:len(token)-1] + "X"
	if _, err := signer.ParseAccess(tampered); err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	signer := platform.NewJWTSigner("supersecret", -1*time.Second)
	token, _, err := signer.SignAccess("user-1", "student", nil, nil)
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}

	_, err = signer.ParseAccess(token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "expir") {
		t.Errorf("expected expiry message, got: %v", err)
	}
}
