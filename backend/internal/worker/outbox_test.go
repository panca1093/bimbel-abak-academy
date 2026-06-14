package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"akademi-bimbel/internal/repository"
)

type mockRepository struct {
	claimOutboxEventsFn       func(context.Context, int) ([]repository.OutboxEvent, error)
	markOutboxProcessedFn     func(context.Context, pgx.Tx, int64) error
	setOrderStatusFn          func(context.Context, pgx.Tx, uuid.UUID, string, string) error
	getOrderByIDFn            func(context.Context, uuid.UUID) (repository.Order, error)
	createCourseEnrollmentFn  func(context.Context, pgx.Tx, repository.CourseEnrollment) error
	createExamRegistrationFn  func(context.Context, pgx.Tx, repository.ExamRegistration) error
	beginTxFn                 func(context.Context) (pgx.Tx, error)
	getExpiredPaymentOrdersFn func(context.Context, int) ([]uuid.UUID, error)
}

func (m *mockRepository) ClaimOutboxEvents(ctx context.Context, limit int) ([]repository.OutboxEvent, error) {
	return m.claimOutboxEventsFn(ctx, limit)
}

func (m *mockRepository) MarkOutboxProcessed(ctx context.Context, tx pgx.Tx, id int64) error {
	return m.markOutboxProcessedFn(ctx, tx, id)
}

func (m *mockRepository) SetOrderStatus(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, status, reason string) error {
	return m.setOrderStatusFn(ctx, tx, orderID, status, reason)
}

func (m *mockRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (repository.Order, error) {
	return m.getOrderByIDFn(ctx, id)
}

func (m *mockRepository) CreateCourseEnrollment(ctx context.Context, tx pgx.Tx, e repository.CourseEnrollment) error {
	return m.createCourseEnrollmentFn(ctx, tx, e)
}

func (m *mockRepository) CreateExamRegistration(ctx context.Context, tx pgx.Tx, reg repository.ExamRegistration) error {
	return m.createExamRegistrationFn(ctx, tx, reg)
}

func (m *mockRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return m.beginTxFn(ctx)
}

func (m *mockRepository) GetExpiredPaymentOrders(ctx context.Context, limit int) ([]uuid.UUID, error) {
	return m.getExpiredPaymentOrdersFn(ctx, limit)
}

type mockTx struct {
	commitFn   func(context.Context) error
	rollbackFn func(context.Context) error
}

func (mt *mockTx) Commit(ctx context.Context) error {
	return mt.commitFn(ctx)
}

func (mt *mockTx) Rollback(ctx context.Context) error {
	return mt.rollbackFn(ctx)
}

// Minimal implementations of unused Tx methods
func (mt *mockTx) Begin(ctx context.Context) (pgx.Tx, error)                               { return nil, nil }
func (mt *mockTx) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error               { return nil }
func (mt *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (mt *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (mt *mockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag(""), nil
}
func (mt *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) { return nil, nil }
func (mt *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row        { return nil }
func (mt *mockTx) Ping(ctx context.Context) error                                        { return nil }
func (mt *mockTx) Conn() *pgx.Conn                                                       { return nil }
func (mt *mockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}
func (mt *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func TestOrderPaidHandlerCreatesEnrollments(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	studentID := uuid.New()
	productID := uuid.New()
	outboxID := int64(1)

	courseEnrollmentCreated := false
	outboxMarkedProcessed := false
	orderStatusUpdated := false

	repo := &mockRepository{
		claimOutboxEventsFn: func(ctx context.Context, limit int) ([]repository.OutboxEvent, error) {
			payload, _ := json.Marshal(OrderPaidPayload{
				OrderID: orderID,
				Items: []OrderItemMini{
					{
						ProductID:   productID,
						ProductType: "course",
					},
				},
			})
			return []repository.OutboxEvent{
				{
					ID:          outboxID,
					AggregateID: orderID,
					EventType:   "OrderPaid",
					Payload:     payload,
					CreatedAt:   time.Now().String(),
				},
			}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (repository.Order, error) {
			return repository.Order{
				ID:        orderID,
				StudentID: studentID,
				Status:    "paid",
			}, nil
		},
		createCourseEnrollmentFn: func(ctx context.Context, tx pgx.Tx, e repository.CourseEnrollment) error {
			if e.StudentID == studentID && e.ProductID == productID && e.OrderID != nil && *e.OrderID == orderID {
				courseEnrollmentCreated = true
			}
			return nil
		},
		createExamRegistrationFn: func(ctx context.Context, tx pgx.Tx, reg repository.ExamRegistration) error {
			return nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			if id == orderID && status == "processing" {
				orderStatusUpdated = true
			}
			return nil
		},
		markOutboxProcessedFn: func(ctx context.Context, tx pgx.Tx, id int64) error {
			if id == outboxID {
				outboxMarkedProcessed = true
			}
			return nil
		},
		beginTxFn: func(ctx context.Context) (pgx.Tx, error) {
			mockTx := &mockTx{
				commitFn:   func(ctx context.Context) error { return nil },
				rollbackFn: func(ctx context.Context) error { return nil },
			}
			return mockTx, nil
		},
	}

	w := &Worker{repo: repo}
	w.pollOutbox(ctx)

	if !courseEnrollmentCreated {
		t.Error("expected CourseEnrollment to be created")
	}
	if !orderStatusUpdated {
		t.Error("expected Order status to be updated to processing")
	}
	if !outboxMarkedProcessed {
		t.Error("expected outbox event to be marked processed")
	}
}

func TestOrderPaidHandlerIdempotent(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	studentID := uuid.New()
	productID := uuid.New()
	outboxID := int64(1)

	createEnrollmentCallCount := 0

	repo := &mockRepository{
		claimOutboxEventsFn: func(ctx context.Context, limit int) ([]repository.OutboxEvent, error) {
			payload, _ := json.Marshal(OrderPaidPayload{
				OrderID: orderID,
				Items: []OrderItemMini{
					{
						ProductID:   productID,
						ProductType: "course",
					},
				},
			})
			return []repository.OutboxEvent{
				{
					ID:          outboxID,
					AggregateID: orderID,
					EventType:   "OrderPaid",
					Payload:     payload,
					CreatedAt:   time.Now().String(),
				},
			}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (repository.Order, error) {
			return repository.Order{
				ID:        orderID,
				StudentID: studentID,
				Status:    "paid",
			}, nil
		},
		createCourseEnrollmentFn: func(ctx context.Context, tx pgx.Tx, e repository.CourseEnrollment) error {
			createEnrollmentCallCount++
			return nil
		},
		createExamRegistrationFn: func(ctx context.Context, tx pgx.Tx, reg repository.ExamRegistration) error {
			return nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			return nil
		},
		markOutboxProcessedFn: func(ctx context.Context, tx pgx.Tx, id int64) error {
			return nil
		},
		beginTxFn: func(ctx context.Context) (pgx.Tx, error) {
			mockTx := &mockTx{
				commitFn:   func(ctx context.Context) error { return nil },
				rollbackFn: func(ctx context.Context) error { return nil },
			}
			return mockTx, nil
		},
	}

	w := &Worker{repo: repo}
	w.pollOutbox(ctx)

	if createEnrollmentCallCount != 1 {
		t.Errorf("expected CreateCourseEnrollment to be called once, got %d", createEnrollmentCallCount)
	}
}

func TestStalePaymentSweeperUpdatesStatus(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	statusUpdated := false

	repo := &mockRepository{
		getExpiredPaymentOrdersFn: func(ctx context.Context, limit int) ([]uuid.UUID, error) {
			return []uuid.UUID{orderID}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (repository.Order, error) {
			return repository.Order{
				ID:        orderID,
				Status:    "payment_pending",
				StudentID: uuid.New(),
				Items: []repository.OrderItem{
					{
						ProductID:   uuid.New(),
						Qty:         2,
						ProductType: "book",
					},
				},
			}, nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			if id == orderID && status == "payment_expired" {
				statusUpdated = true
			}
			return nil
		},
		beginTxFn: func(ctx context.Context) (pgx.Tx, error) {
			mockTx := &mockTx{
				commitFn:   func(ctx context.Context) error { return nil },
				rollbackFn: func(ctx context.Context) error { return nil },
			}
			return mockTx, nil
		},
	}

	w := &Worker{repo: repo, sweeperInterval: time.Minute}
	w.sweepStalePayments(ctx)

	if !statusUpdated {
		t.Error("expected Order status to be updated to payment_expired")
	}
}

func TestStalePaymentSweeperIdempotent(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	checkStatusCallCount := 0

	repo := &mockRepository{
		getExpiredPaymentOrdersFn: func(ctx context.Context, limit int) ([]uuid.UUID, error) {
			return []uuid.UUID{orderID}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (repository.Order, error) {
			return repository.Order{
				ID:        orderID,
				Status:    "payment_pending",
				StudentID: uuid.New(),
				Items: []repository.OrderItem{
					{
						ProductID:   uuid.New(),
						Qty:         1,
						ProductType: "book",
					},
				},
			}, nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			checkStatusCallCount++
			return nil
		},
		beginTxFn: func(ctx context.Context) (pgx.Tx, error) {
			mockTx := &mockTx{
				commitFn:   func(ctx context.Context) error { return nil },
				rollbackFn: func(ctx context.Context) error { return nil },
			}
			return mockTx, nil
		},
	}

	w := &Worker{repo: repo, sweeperInterval: time.Minute}
	w.sweepStalePayments(ctx)

	if checkStatusCallCount == 0 {
		t.Error("expected SetOrderStatus to be called at least once")
	}
}
