package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID                    uuid.UUID
	StudentID             uuid.UUID
	Status                string
	Subtotal              float64
	Discount              float64
	ShippingCost          float64
	Total                 float64
	PromoCodeID           *uuid.UUID
	ShippingAddress       []byte
	SelectedCourier       string
	TrackingNumber        string
	ShippedAt             *time.Time
	GatewayRef            string
	PaymentMethod         string
	PaymentExpiresAt      *time.Time
	PaidAt                *time.Time
	InvoiceURL            string
	EstimatedDeliveryDays string
	CheckedOutAt          *time.Time
	CompletedAt           *time.Time
	CancelledAt           *time.Time
	CancellationReason    string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	Items                 []OrderItem
}

type OrderItem struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	ProductType string
	Name        string
	UnitPrice   float64
	Qty         int
	Jumlah      float64
	WeightGrams int
	FulfilledAt *time.Time
	CreatedAt   time.Time
}
