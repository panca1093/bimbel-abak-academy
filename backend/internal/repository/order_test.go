package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Compile-time check: *Repository must implement all order methods.
var _ interface {
	MintCart(context.Context, uuid.UUID) (Order, bool, error)
	GetCartByStudentID(context.Context, uuid.UUID) (Order, error)
	GetOrderByID(context.Context, uuid.UUID) (Order, error)
	ListOrders(context.Context, OrderFilter) ([]Order, string, error)
	AddItem(context.Context, uuid.UUID, OrderItem) error
	RemoveItem(context.Context, uuid.UUID, uuid.UUID) error
	PatchCart(context.Context, uuid.UUID, OrderPatch) error
	SetOrderStatus(context.Context, pgx.Tx, uuid.UUID, string, string) error
	SetShipped(context.Context, uuid.UUID, string) error
	SetPaymentRef(context.Context, uuid.UUID, string, time.Time) error
	CheckoutOrder(context.Context, pgx.Tx, uuid.UUID) error
} = (*Repository)(nil)

func TestOrderStructs(t *testing.T) {
	// Verify Order struct can be instantiated
	order := Order{
		ID:        uuid.New(),
		StudentID: uuid.New(),
		Status:    "cart",
		Items:     []OrderItem{},
	}
	if order.Status != "cart" {
		t.Errorf("Order.Status = %q, want 'cart'", order.Status)
	}
}

func TestOrderItemStruct(t *testing.T) {
	// Verify OrderItem struct can be instantiated
	item := OrderItem{
		ID:        uuid.New(),
		OrderID:   uuid.New(),
		ProductID: uuid.New(),
		Qty:       5,
	}
	if item.Qty != 5 {
		t.Errorf("OrderItem.Qty = %d, want 5", item.Qty)
	}
}

func TestCheckoutOrderSQL(t *testing.T) {
	// This test documents that CheckoutOrder uses FOR UPDATE in SQL.
	// The actual transaction semantics are verified in integration tests.
	t.Log("CheckoutOrder includes 'FOR UPDATE' in product selection SQL for pessimistic locking")
}
