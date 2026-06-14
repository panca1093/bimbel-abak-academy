package model

import (
	"time"

	"github.com/google/uuid"
)

type PromoCode struct {
	ID                uuid.UUID
	Code              string
	DiscountPercent   *float64
	DiscountAmount    *float64
	MinOrderAmount    *float64
	MaxDiscountAmount *float64
	MaxUses           *int
	Uses              int
	ExpiresAt         *time.Time
	CreatedAt         time.Time
}
