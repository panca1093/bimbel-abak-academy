package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPromoCodeMethods(t *testing.T) {
	// Compile-time test: verify all PromoCode methods exist on *Repository

	r := &Repository{}
	ctx := context.Background()

	// GetPromoByCode
	_ = r.GetPromoByCode

	// CreatePromoCode
	_ = r.CreatePromoCode

	// UpdatePromoCode
	_ = r.UpdatePromoCode

	// DeletePromoCode
	_ = r.DeletePromoCode

	// ListPromoCodes
	_ = r.ListPromoCodes

	// IncrementPromoUses
	_ = r.IncrementPromoUses

	// Verify method signatures compile
	promo := PromoCode{
		ID:              uuid.New(),
		Code:            "TEST",
		DiscountPercent: ptrFloat64(10.0),
		Uses:            0,
		CreatedAt:       time.Now(),
	}
	_ = promo

	_ = ctx
}

func ptrFloat64(v float64) *float64 {
	return &v
}
