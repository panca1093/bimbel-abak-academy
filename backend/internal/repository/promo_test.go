package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
)

func TestPromoCodeMethods(t *testing.T) {
	r := &Repository{}
	ctx := context.Background()

	_ = r.GetPromoByCode
	_ = r.CreatePromoCode
	_ = r.UpdatePromoCode
	_ = r.DeletePromoCode
	_ = r.ListPromoCodes
	_ = r.IncrementPromoUses

	promo := model.PromoCode{
		ID:              uuid.New(),
		Code:            "TEST",
		DiscountPercent: ptrFloat64(10.0),
		UsedCount:       0,
		CreatedAt:       time.Now(),
	}
	_ = promo

	_ = ctx
}

func ptrFloat64(v float64) *float64 {
	return &v
}
