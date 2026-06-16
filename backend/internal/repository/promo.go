package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
)

func (r *Repository) GetPromoByCode(ctx context.Context, code string) (model.PromoCode, error) {
	p := model.PromoCode{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, code, discount_percent, discount_amount, min_order_amount, max_discount_amount, max_uses, used_count, expires_at, created_at
		FROM promo_code
		WHERE code = $1`,
		code,
	).Scan(
		&p.ID, &p.Code, &p.DiscountPercent, &p.DiscountAmount, &p.MinOrderAmount, &p.MaxDiscountAmount, &p.MaxUses, &p.UsedCount, &p.ExpiresAt, &p.CreatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return model.PromoCode{}, nil
		}
		return model.PromoCode{}, err
	}
	return p, nil
}

func (r *Repository) CreatePromoCode(ctx context.Context, p model.PromoCode) (model.PromoCode, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO promo_code (code, discount_percent, discount_amount, min_order_amount, max_discount_amount, max_uses, used_count, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, code, discount_percent, discount_amount, min_order_amount, max_discount_amount, max_uses, used_count, expires_at, created_at`,
		p.Code, p.DiscountPercent, p.DiscountAmount, p.MinOrderAmount, p.MaxDiscountAmount, p.MaxUses, p.UsedCount, p.ExpiresAt,
	).Scan(
		&p.ID, &p.Code, &p.DiscountPercent, &p.DiscountAmount, &p.MinOrderAmount, &p.MaxDiscountAmount, &p.MaxUses, &p.UsedCount, &p.ExpiresAt, &p.CreatedAt,
	)
	return p, err
}

func (r *Repository) UpdatePromoCode(ctx context.Context, id uuid.UUID, maxUses *int, expiresAt *time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE promo_code SET max_uses = $1, expires_at = $2 WHERE id = $3`,
		maxUses, expiresAt, id,
	)
	return err
}

func (r *Repository) DeletePromoCode(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM promo_code WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) ListPromoCodes(ctx context.Context) ([]model.PromoCode, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, code, discount_percent, discount_amount, min_order_amount, max_discount_amount, max_uses, used_count, expires_at, created_at
		FROM promo_code
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promos []model.PromoCode
	for rows.Next() {
		p := model.PromoCode{}
		err := rows.Scan(
			&p.ID, &p.Code, &p.DiscountPercent, &p.DiscountAmount, &p.MinOrderAmount, &p.MaxDiscountAmount, &p.MaxUses, &p.UsedCount, &p.ExpiresAt, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}
	return promos, rows.Err()
}

func (r *Repository) IncrementPromoUses(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE promo_code SET used_count = used_count + 1 WHERE id = $1`,
		id,
	)
	return err
}
