package repository

import (
	"context"
)

// Compile-time check: *Repository must implement all product methods.
var _ interface {
	CreateProduct(context.Context, *Product) error
	GetProductByID(context.Context, string) (*Product, error)
	ListProducts(context.Context, ProductFilter) ([]Product, string, error)
	UpdateProduct(context.Context, string, *Product) error
	PublishProduct(context.Context, string) error
	DeleteProduct(context.Context, string) error
	ArchiveProduct(context.Context, string) error
} = (*Repository)(nil)
