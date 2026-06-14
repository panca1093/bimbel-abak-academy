package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID                 uuid.UUID
	StudentID          uuid.UUID
	Status             string
	Subtotal           float64
	Discount           float64
	ShippingAmount     float64
	Total              float64
	PromoCodeID        *uuid.UUID
	ShippingAddress    []byte
	Courier            string
	TrackingNumber     string
	ShippedAt          *time.Time
	PaymentRef         string
	PaymentExpiresAt   *time.Time
	CancellationReason string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Items              []OrderItem
}

type OrderItem struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	ProductType string
	Title       string
	UnitPrice   float64
	Qty         int
	FulfilledAt *time.Time
	CreatedAt   time.Time
}
