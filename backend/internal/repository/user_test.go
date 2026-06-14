package repository

import (
	"context"
	"testing"

	"akademi-bimbel/internal/model"
)

// Compile-time check: *Repository must implement all user methods.
var _ interface {
	CreateUser(context.Context, *model.User) error
	GetUserByEmail(context.Context, string) (*model.User, error)
	GetUserByID(context.Context, string) (*model.User, error)
	UpdatePasswordHash(context.Context, string, string) error
	TombstoneUser(context.Context, string) error
} = (*Repository)(nil)

func TestNormalizeEmail(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"user@example.com", "user@example.com"},
		{"User@Example.COM", "user@example.com"},
		{"ADMIN@AMARTHA.COM", "admin@amartha.com"},
		{"", ""},
	}
	for _, c := range cases {
		got := normalizeEmail(c.in)
		if got != c.want {
			t.Errorf("normalizeEmail(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
