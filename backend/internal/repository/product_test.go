package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
)

// Compile-time check: *Repository must implement all product methods.
var _ interface {
	CreateProduct(context.Context, *model.Product) error
	GetProductByID(context.Context, string) (*model.Product, error)
	ListProducts(context.Context, ProductFilter) ([]model.Product, string, error)
	UpdateProduct(context.Context, string, *model.Product) error
	PublishProduct(context.Context, string) error
	DeleteProduct(context.Context, string) error
	ArchiveProduct(context.Context, string) error
	CreateProductWithCourses(context.Context, pgx.Tx, *model.Product, []uuid.UUID) error
} = (*Repository)(nil)
