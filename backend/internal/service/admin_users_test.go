package service

import (
	"context"
	"errors"
	"testing"

	"akademi-bimbel/internal/model"
)

func TestIsValidAdminRole(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{RoleSuperAdmin, true},
		{RoleAdminStore, true},
		{RoleAdminExam, true},
		{RoleAdminSchool, true},
		{"student", false},
		{"ghost", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.role, func(t *testing.T) {
			got := isValidAdminRole(tc.role)
			if got != tc.want {
				t.Errorf("isValidAdminRole(%q): want %v, got %v", tc.role, tc.want, got)
			}
		})
	}
}

func TestIsValidStatusFilter(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"active", true},
		{"deactivated", true},
		{"deleted", false},
		{"pending", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.status, func(t *testing.T) {
			got := isValidStatusFilter(tc.status)
			if got != tc.want {
				t.Errorf("isValidStatusFilter(%q): want %v, got %v", tc.status, tc.want, got)
			}
		})
	}
}

// -- self-deactivation guard --

func TestCheckSelfDeactivation(t *testing.T) {
	t.Run("same user self-deactivation returns ErrCannotDeactivateSelf", func(t *testing.T) {
		err := checkSelfDeactivation("u1", "u1", "deactivated")
		if !errors.Is(err, ErrCannotDeactivateSelf) {
			t.Errorf("want ErrCannotDeactivateSelf, got %v", err)
		}
	})

	t.Run("different user deactivation returns nil", func(t *testing.T) {
		err := checkSelfDeactivation("u2", "u1", "deactivated")
		if err != nil {
			t.Errorf("want nil, got %v", err)
		}
	})

	t.Run("self activation is allowed", func(t *testing.T) {
		err := checkSelfDeactivation("u1", "u1", "active")
		if err != nil {
			t.Errorf("want nil, got %v", err)
		}
	})

	t.Run("self status change to non-deactivated returns nil", func(t *testing.T) {
		err := checkSelfDeactivation("u1", "u1", "active")
		if err != nil {
			t.Errorf("want nil, got %v", err)
		}
	})
}

// -- email uniqueness --

func TestCheckEmailUniqueness(t *testing.T) {
	ctx := context.Background()

	t.Run("taken email returns ErrEmailTaken", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&model.User{
			Email:  strptr("taken@example.com"),
			Status: "active",
			Role:   RoleSuperAdmin,
		})
		err := checkEmailUniqueness(ctx, repo, "taken@example.com")
		if !errors.Is(err, ErrEmailTaken) {
			t.Errorf("want ErrEmailTaken, got %v", err)
		}
	})

	t.Run("available email returns nil", func(t *testing.T) {
		repo := newFakeUserRepo()
		err := checkEmailUniqueness(ctx, repo, "new@example.com")
		if err != nil {
			t.Errorf("want nil, got %v", err)
		}
	})

	t.Run("deleted user does not block email", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&model.User{
			Email:  strptr("deleted@example.com"),
			Status: "deleted",
			Role:   RoleSuperAdmin,
		})
		err := checkEmailUniqueness(ctx, repo, "deleted@example.com")
		if err != nil {
			t.Errorf("deleted user should not block email, got %v", err)
		}
	})

	t.Run("case-insensitive match returns ErrEmailTaken", func(t *testing.T) {
		repo := newFakeUserRepo()
		repo.seed(&model.User{
			Email:  strptr("Case@Example.com"),
			Status: "active",
			Role:   RoleSuperAdmin,
		})
		// Different case should match normalized
		err := checkEmailUniqueness(ctx, repo, "case@example.com")
		if !errors.Is(err, ErrEmailTaken) {
			t.Errorf("want ErrEmailTaken for case-insensitive match, got %v", err)
		}
	})
}
