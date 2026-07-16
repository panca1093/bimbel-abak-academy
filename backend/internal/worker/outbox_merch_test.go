package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
)

// A paid order containing a merchandise item must be classified physical exactly
// like a book: it lands in "processing" (ship-pending), is NOT auto-completed, and
// no course-session / exam-registration entitlement is provisioned for it.
func TestOrderPaidHandler_MerchandiseItem_ShipPendingNoEntitlement(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	productID := uuid.New()
	outboxID := int64(1)

	var statusUpdate string
	coursesLookedUp := false
	examsLookedUp := false
	sessionCreated := false
	registrationCreated := false

	repo := &mockRepository{
		claimOutboxEventsFn: func(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
			payload, _ := json.Marshal(OrderPaidPayload{
				OrderID: orderID,
				Items:   []OrderItemMini{{ProductID: productID, ProductType: "merchandise"}},
			})
			return []model.OutboxEvent{
				{ID: outboxID, AggregateID: orderID, EventType: "OrderPaid", Payload: payload, CreatedAt: time.Now().String()},
			}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (model.Order, error) {
			return model.Order{ID: orderID, StudentID: uuid.New(), Status: "paid"}, nil
		},
		getCoursesByProductIDFn: func(ctx context.Context, pid uuid.UUID) ([]model.Course, error) {
			coursesLookedUp = true
			return nil, nil
		},
		getExamsByProductIDFn: func(ctx context.Context, pid uuid.UUID) ([]model.Exam, error) {
			examsLookedUp = true
			return nil, nil
		},
		createCourseSessionFn: func(ctx context.Context, tx pgx.Tx, s model.CourseSession) error {
			sessionCreated = true
			return nil
		},
		createExamRegistrationFn: func(ctx context.Context, tx pgx.Tx, reg model.ExamRegistration) error {
			registrationCreated = true
			return nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			statusUpdate = status
			return nil
		},
		markOutboxProcessedFn: func(ctx context.Context, tx pgx.Tx, id int64) error { return nil },
		beginTxFn: func(ctx context.Context) (pgx.Tx, error) {
			return &mockTx{commitFn: func(ctx context.Context) error { return nil }, rollbackFn: func(ctx context.Context) error { return nil }}, nil
		},
	}

	w := &Worker{repo: repo}
	w.pollOutbox(ctx)

	if statusUpdate != "processing" {
		t.Errorf("merchandise order status = %q, want %q (physical → ship-pending, not auto-completed)", statusUpdate, "processing")
	}
	if coursesLookedUp || examsLookedUp {
		t.Error("merchandise must not trigger course/exam entitlement lookup")
	}
	if sessionCreated || registrationCreated {
		t.Error("merchandise must not provision a course-session or exam-registration")
	}
}
