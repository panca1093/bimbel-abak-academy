package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
)

var ErrNotFound = errors.New("not found")

type ProductFilter struct {
	Type          string
	Status        string
	IsVisibleOnly bool
	Cursor        string
	Limit         int
}

func (r *Repository) CreateProduct(ctx context.Context, p *model.Product) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO product (type, title, description, price, stock, status, is_visible, weight_grams, cover_image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`,
		p.Type, p.Title, p.Description, p.Price, p.Stock, p.Status, p.IsVisible, p.WeightGrams, p.CoverImageURL,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	return err
}

func (r *Repository) GetProductByID(ctx context.Context, id string) (*model.Product, error) {
	p := &model.Product{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, type, title, description, price, stock, status, is_visible, weight_grams, cover_image_url, created_at, updated_at
		FROM product
		WHERE id = $1`,
		id,
	).Scan(
		&p.ID, &p.Type, &p.Title, &p.Description, &p.Price, &p.Stock, &p.Status, &p.IsVisible, &p.WeightGrams, &p.CoverImageURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *Repository) ListProducts(ctx context.Context, filter ProductFilter) ([]model.Product, string, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	var products []model.Product
	query := `SELECT id, type, title, description, price, stock, status, is_visible, weight_grams, cover_image_url, created_at, updated_at
	FROM product WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.Type != "" {
		query += fmt.Sprintf(` AND type = $%d`, argIdx)
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.IsVisibleOnly {
		query += fmt.Sprintf(` AND is_visible = $%d`, argIdx)
		args = append(args, true)
		argIdx++
	}
	if filter.Cursor != "" {
		query += fmt.Sprintf(` AND id > $%d`, argIdx)
		args = append(args, filter.Cursor)
		argIdx++
	}

	query += ` ORDER BY id LIMIT $` + fmt.Sprintf("%d", argIdx)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		p := model.Product{}
		err := rows.Scan(
			&p.ID, &p.Type, &p.Title, &p.Description, &p.Price, &p.Stock, &p.Status, &p.IsVisible, &p.WeightGrams, &p.CoverImageURL, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(products) > filter.Limit {
		nextCursor = products[filter.Limit].ID
		products = products[:filter.Limit]
	}

	return products, nextCursor, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, id string, p *model.Product) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE product
		SET type = $1, title = $2, description = $3, price = $4, stock = $5, status = $6, is_visible = $7, weight_grams = $8, cover_image_url = $9, updated_at = now()
		WHERE id = $10`,
		p.Type, p.Title, p.Description, p.Price, p.Stock, p.Status, p.IsVisible, p.WeightGrams, p.CoverImageURL, id,
	)
	return err
}

func (r *Repository) PublishProduct(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE product SET status = 'published', updated_at = now() WHERE id = $1 AND status = 'draft'`,
		id,
	)
	return err
}

func (r *Repository) DeleteProduct(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM product WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) ArchiveProduct(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE product SET status = 'archived', updated_at = now() WHERE id = $1`,
		id,
	)
	return err
}

// ReplaceProductCourses atomically replaces all product_course links for a product.
func (r *Repository) ReplaceProductCourses(ctx context.Context, tx pgx.Tx, productID uuid.UUID, courseIDs []uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM product_course WHERE product_id = $1`, productID)
	if err != nil {
		return err
	}
	for _, courseID := range courseIDs {
		_, err := tx.Exec(ctx,
			`INSERT INTO product_course (product_id, course_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			productID, courseID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateProductWithCourses inserts a product and its product_course links in one transaction.
func (r *Repository) CreateProductWithCourses(ctx context.Context, tx pgx.Tx, p *model.Product, courseIDs []uuid.UUID) error {
	err := tx.QueryRow(ctx,
		`INSERT INTO product (type, title, description, price, stock, status, is_visible, weight_grams, cover_image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`,
		p.Type, p.Title, p.Description, p.Price, p.Stock, p.Status, p.IsVisible, p.WeightGrams, p.CoverImageURL,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	for _, courseID := range courseIDs {
		_, err := tx.Exec(ctx,
			`INSERT INTO product_course (product_id, course_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING`,
			p.ID, courseID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
