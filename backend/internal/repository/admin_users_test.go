package repository

import (
	"context"

	"akademi-bimbel/internal/model"
)

// Compile-time check: *Repository must implement all admin user methods.
var _ interface {
	ListAdminUsers(context.Context, AdminUserFilter) ([]AdminUserRow, string, error)
	CreateAdminUser(context.Context, *model.User) error
	GetAdminUserByID(context.Context, string) (*model.User, error)
	UpdateAdminUserRole(context.Context, string, string) error
	UpdateAdminUserStatus(context.Context, string, string) error
} = (*Repository)(nil)
