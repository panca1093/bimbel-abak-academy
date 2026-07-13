package integration_test

import (
	"context"
	"testing"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"

	"github.com/stretchr/testify/require"
)

// TestAuthProviderRoundTrip ensures the new auth_provider column is threaded
// through CreateUser + every full-column GetUserBy* scan site.
func TestAuthProviderRoundTrip(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	repo := repository.New(env.pool)

	email := "auth-provider-test@example.com"
	u := &model.User{
		Email:        &email,
		Username:     strPtr("authprov"),
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		Role:         "student",
		Name:         "Auth Provider Test",
		Status:       "active",
		OTPEnabled:   false,
	}
	require.NoError(t, repo.CreateUser(ctx, u))
	require.NotEmpty(t, u.ID)

	t.Run("round-trip defaults to password", func(t *testing.T) {
		got, err := repo.GetUserByID(ctx, u.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "password", got.AuthProvider)
	})

	t.Run("GetUserByEmail carries auth_provider", func(t *testing.T) {
		got, err := repo.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "password", got.AuthProvider)
	})

	t.Run("GetUserByUsername carries auth_provider", func(t *testing.T) {
		got, err := repo.GetUserByUsername(ctx, "authprov")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "password", got.AuthProvider)
	})
}

func strPtr(s string) *string { return &s }
