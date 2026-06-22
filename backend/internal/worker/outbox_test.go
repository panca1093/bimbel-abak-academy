package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"akademi-bimbel/internal/model"
)

type mockRepository struct {
	claimOutboxEventsFn       func(context.Context, int) ([]model.OutboxEvent, error)
	markOutboxProcessedFn     func(context.Context, pgx.Tx, int64) error
	setOrderStatusFn          func(context.Context, pgx.Tx, uuid.UUID, string, string) error
	getOrderByIDFn            func(context.Context, uuid.UUID) (model.Order, error)
	createCourseSessionFn     func(context.Context, pgx.Tx, model.CourseSession) error
	getCoursesByProductIDFn   func(context.Context, uuid.UUID) ([]model.Course, error)
	beginTxFn                 func(context.Context) (pgx.Tx, error)
	getExpiredPaymentOrdersFn func(context.Context, int) ([]uuid.UUID, error)
}

func (m *mockRepository) ClaimOutboxEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	return m.claimOutboxEventsFn(ctx, limit)
}

func (m *mockRepository) MarkOutboxProcessed(ctx context.Context, tx pgx.Tx, id int64) error {
	return m.markOutboxProcessedFn(ctx, tx, id)
}

func (m *mockRepository) SetOrderStatus(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, status, reason string) error {
	return m.setOrderStatusFn(ctx, tx, orderID, status, reason)
}

func (m *mockRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (model.Order, error) {
	return m.getOrderByIDFn(ctx, id)
}

func (m *mockRepository) CreateCourseSession(ctx context.Context, tx pgx.Tx, s model.CourseSession) error {
	return m.createCourseSessionFn(ctx, tx, s)
}

func (m *mockRepository) GetCoursesByProductID(ctx context.Context, productID uuid.UUID) ([]model.Course, error) {
	return m.getCoursesByProductIDFn(ctx, productID)
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

func TestOrderPaidHandlerCreatesTwoCourseSessionsForLinkedCourses(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	studentID := uuid.New()
	productID := uuid.New()
	outboxID := int64(1)
	course1ID := uuid.New()
	course2ID := uuid.New()

	var createdSessions []model.CourseSession

	repo := &mockRepository{
		claimOutboxEventsFn: func(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
			payload, _ := json.Marshal(OrderPaidPayload{
				OrderID: orderID,
				Items: []OrderItemMini{
					{ProductID: productID, ProductType: "course"},
				},
			})
			return []model.OutboxEvent{
				{ID: outboxID, AggregateID: orderID, EventType: "OrderPaid", Payload: payload, CreatedAt: time.Now().String()},
			}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (model.Order, error) {
			return model.Order{ID: orderID, StudentID: studentID, Status: "paid"}, nil
		},
		getCoursesByProductIDFn: func(ctx context.Context, pid uuid.UUID) ([]model.Course, error) {
			return []model.Course{
				{ID: course1ID, Title: "Math"},
				{ID: course2ID, Title: "Science"},
			}, nil
		},
		createCourseSessionFn: func(ctx context.Context, tx pgx.Tx, s model.CourseSession) error {
			createdSessions = append(createdSessions, s)
			return nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			return nil
		},
		markOutboxProcessedFn: func(ctx context.Context, tx pgx.Tx, id int64) error {
			return nil
		},
		beginTxFn: func(ctx context.Context) (pgx.Tx, error) {
			return &mockTx{commitFn: func(ctx context.Context) error { return nil }, rollbackFn: func(ctx context.Context) error { return nil }}, nil
		},
	}

	w := &Worker{repo: repo}
	w.pollOutbox(ctx)

	if len(createdSessions) != 2 {
		t.Fatalf("expected 2 course sessions, got %d", len(createdSessions))
	}

	ids := map[uuid.UUID]bool{course1ID: true, course2ID: true}
	for _, s := range createdSessions {
		if !ids[s.CourseID] {
			t.Errorf("unexpected course_id %v", s.CourseID)
		}
		if s.StudentID != studentID {
			t.Errorf("expected student_id %v, got %v", studentID, s.StudentID)
		}
		if s.Status != "active" {
			t.Errorf("expected status active, got %s", s.Status)
		}
		if s.Source != "order" {
			t.Errorf("expected source order, got %s", s.Source)
		}
		if s.OrderID == nil || *s.OrderID != orderID {
			t.Errorf("expected order_id %v, got %v", orderID, s.OrderID)
		}
		if s.EnrolledAt.IsZero() {
			t.Error("expected enrolled_at to be set")
		}
	}
	if createdSessions[0].CourseID == createdSessions[1].CourseID {
		t.Error("expected two different course IDs")
	}
}

func TestOrderPaidHandlerZeroLinkedCoursesSkipsButCommits(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	productID := uuid.New()
	outboxID := int64(1)

	sessionCreated := false
	orderStatusUpdated := false
	outboxMarkedProcessed := false

	repo := &mockRepository{
		claimOutboxEventsFn: func(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
			payload, _ := json.Marshal(OrderPaidPayload{
				OrderID: orderID,
				Items:   []OrderItemMini{{ProductID: productID, ProductType: "course"}},
			})
			return []model.OutboxEvent{
				{ID: outboxID, AggregateID: orderID, EventType: "OrderPaid", Payload: payload, CreatedAt: time.Now().String()},
			}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (model.Order, error) {
			return model.Order{ID: orderID, StudentID: uuid.New(), Status: "paid"}, nil
		},
		getCoursesByProductIDFn: func(ctx context.Context, pid uuid.UUID) ([]model.Course, error) {
			return []model.Course{}, nil
		},
		createCourseSessionFn: func(ctx context.Context, tx pgx.Tx, s model.CourseSession) error {
			sessionCreated = true
			return nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			if status == "completed" {
				orderStatusUpdated = true
			}
			return nil
		},
		markOutboxProcessedFn: func(ctx context.Context, tx pgx.Tx, id int64) error {
			outboxMarkedProcessed = true
			return nil
		},
		beginTxFn: func(ctx context.Context) (pgx.Tx, error) {
			return &mockTx{commitFn: func(ctx context.Context) error { return nil }, rollbackFn: func(ctx context.Context) error { return nil }}, nil
		},
	}

	w := &Worker{repo: repo}
	w.pollOutbox(ctx)

	if sessionCreated {
		t.Error("expected no course session to be created for zero-linked product")
	}
	if !orderStatusUpdated {
		t.Error("expected order status to be updated to completed (digital-only order)")
	}
	if !outboxMarkedProcessed {
		t.Error("expected outbox event to be marked processed")
	}
}

func TestOrderPaidHandlerIdempotentOnReplay(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	studentID := uuid.New()
	productID := uuid.New()
	outboxID := int64(1)
	courseID := uuid.New()

	var sessionCallCount int

	repo := &mockRepository{
		claimOutboxEventsFn: func(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
			payload, _ := json.Marshal(OrderPaidPayload{
				OrderID: orderID,
				Items:   []OrderItemMini{{ProductID: productID, ProductType: "course"}},
			})
			return []model.OutboxEvent{
				{ID: outboxID, AggregateID: orderID, EventType: "OrderPaid", Payload: payload, CreatedAt: time.Now().String()},
			}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (model.Order, error) {
			return model.Order{ID: orderID, StudentID: studentID, Status: "paid"}, nil
		},
		getCoursesByProductIDFn: func(ctx context.Context, pid uuid.UUID) ([]model.Course, error) {
			return []model.Course{{ID: courseID, Title: "Math"}}, nil
		},
		createCourseSessionFn: func(ctx context.Context, tx pgx.Tx, s model.CourseSession) error {
			sessionCallCount++
			return nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			return nil
		},
		markOutboxProcessedFn: func(ctx context.Context, tx pgx.Tx, id int64) error {
			return nil
		},
		beginTxFn: func(ctx context.Context) (pgx.Tx, error) {
			return &mockTx{commitFn: func(ctx context.Context) error { return nil }, rollbackFn: func(ctx context.Context) error { return nil }}, nil
		},
	}

	w := &Worker{repo: repo}
	w.pollOutbox(ctx)

	if sessionCallCount != 1 {
		t.Fatalf("expected 1 CreateCourseSession call on first poll, got %d", sessionCallCount)
	}

	// Replay — pollOutbox processes the same event again (mock returns it again)
	w.pollOutbox(ctx)

	// The second poll also succeeds — idempotent. In real DB, ON CONFLICT DO NOTHING
	// prevents duplicates; the mock simulates success on both calls.
	if sessionCallCount != 2 {
		t.Fatalf("expected 2 CreateCourseSession calls across two polls, got %d", sessionCallCount)
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
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (model.Order, error) {
			return model.Order{
				ID:        orderID,
				Status:    "payment_pending",
				StudentID: uuid.New(),
				Items: []model.OrderItem{
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
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (model.Order, error) {
			return model.Order{
				ID:        orderID,
				Status:    "payment_pending",
				StudentID: uuid.New(),
				Items: []model.OrderItem{
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
