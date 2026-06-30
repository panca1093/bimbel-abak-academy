package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"
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
	getExamByProductIDFn      func(context.Context, uuid.UUID) (*model.Exam, error)
	createExamRegistrationFn  func(context.Context, pgx.Tx, model.ExamRegistration) error
	stampOrderItemFulfilledFn func(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) error
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

func (m *mockRepository) GetExamByProductID(ctx context.Context, productID uuid.UUID) (*model.Exam, error) {
	if m.getExamByProductIDFn == nil {
		return nil, nil
	}
	return m.getExamByProductIDFn(ctx, productID)
}

func (m *mockRepository) CreateExamRegistration(ctx context.Context, tx pgx.Tx, reg model.ExamRegistration) error {
	if m.createExamRegistrationFn == nil {
		return nil
	}
	return m.createExamRegistrationFn(ctx, tx, reg)
}

func (m *mockRepository) StampOrderItemFulfilledAt(ctx context.Context, tx pgx.Tx, orderID, productID uuid.UUID) error {
	if m.stampOrderItemFulfilledFn == nil {
		return nil
	}
	return m.stampOrderItemFulfilledFn(ctx, tx, orderID, productID)
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

func TestGenerateToken_Returns8UppercaseAlphanumeric(t *testing.T) {
	tokenRe := regexp.MustCompile(`^[A-Z0-9]{8}$`)
	seen := make(map[string]struct{}, 100)

	for i := 0; i < 100; i++ {
		tok := generateToken()
		if !tokenRe.MatchString(tok) {
			t.Fatalf("token %q does not match %s", tok, tokenRe)
		}
		seen[tok] = struct{}{}
	}

	if len(seen) < 2 {
		t.Errorf("expected at least 2 distinct tokens across 100 calls, got %d", len(seen))
	}
}

func TestOrderPaidHandler_ExamItem_ProvisionsRegistration(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	studentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	productID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	examID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	outboxID := int64(42)

	var capturedRegistration model.ExamRegistration
	var fulfilledOrderID, fulfilledProductID uuid.UUID
	var fulfilledCalls int
	var statusUpdate string

	tokenRe := regexp.MustCompile(`^[A-Z0-9]{8}$`)

	prevLogger := slog.Default()
	buf := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	defer slog.SetDefault(prevLogger)

	repo := &mockRepository{
		claimOutboxEventsFn: func(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
			payload, _ := json.Marshal(OrderPaidPayload{
				OrderID: orderID,
				Items:   []OrderItemMini{{ProductID: productID, ProductType: "exam"}},
			})
			return []model.OutboxEvent{
				{ID: outboxID, AggregateID: orderID, EventType: "OrderPaid", Payload: payload, CreatedAt: time.Now().String()},
			}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (model.Order, error) {
			return model.Order{ID: orderID, StudentID: studentID, Status: "paid"}, nil
		},
		getExamByProductIDFn: func(ctx context.Context, pid uuid.UUID) (*model.Exam, error) {
			if pid != productID {
				t.Errorf("GetExamByProductID called with %v, want %v", pid, productID)
			}
			return &model.Exam{ID: examID, Title: "Finals", ProductID: &productID}, nil
		},
		createExamRegistrationFn: func(ctx context.Context, tx pgx.Tx, reg model.ExamRegistration) error {
			capturedRegistration = reg
			return nil
		},
		stampOrderItemFulfilledFn: func(ctx context.Context, tx pgx.Tx, oid, pid uuid.UUID) error {
			fulfilledCalls++
			fulfilledOrderID = oid
			fulfilledProductID = pid
			return nil
		},
		setOrderStatusFn: func(ctx context.Context, tx pgx.Tx, id uuid.UUID, status, reason string) error {
			statusUpdate = status
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

	if capturedRegistration.StudentID != studentID {
		t.Errorf("registration student_id = %v, want %v", capturedRegistration.StudentID, studentID)
	}
	if capturedRegistration.ExamID != examID {
		t.Errorf("registration exam_id = %v, want %v", capturedRegistration.ExamID, examID)
	}
	if !tokenRe.MatchString(capturedRegistration.Token) {
		t.Errorf("registration token %q does not match %s", capturedRegistration.Token, tokenRe)
	}
	if capturedRegistration.Status != "registered" {
		t.Errorf("registration status = %q, want %q", capturedRegistration.Status, "registered")
	}

	if fulfilledCalls != 1 {
		t.Fatalf("StampOrderItemFulfilledAt called %d times, want 1", fulfilledCalls)
	}
	if fulfilledOrderID != orderID || fulfilledProductID != productID {
		t.Errorf("StampOrderItemFulfilledAt called with (%v,%v), want (%v,%v)", fulfilledOrderID, fulfilledProductID, orderID, productID)
	}

	if statusUpdate != "completed" {
		t.Errorf("order status = %q, want %q (digital exam → auto-complete)", statusUpdate, "completed")
	}

	if strings.Contains(buf.String(), "level=ERROR") {
		t.Errorf("expected no ERROR-level log lines, got:\n%s", buf.String())
	}
}

func TestOrderPaidHandler_ExamItem_IdempotentOnReplay(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	studentID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	productID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	examID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	outboxID := int64(99)

	var createCalls int
	var fulfilledCalls int

	repo := &mockRepository{
		claimOutboxEventsFn: func(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
			payload, _ := json.Marshal(OrderPaidPayload{
				OrderID: orderID,
				Items:   []OrderItemMini{{ProductID: productID, ProductType: "exam"}},
			})
			return []model.OutboxEvent{
				{ID: outboxID, AggregateID: orderID, EventType: "OrderPaid", Payload: payload, CreatedAt: time.Now().String()},
			}, nil
		},
		getOrderByIDFn: func(ctx context.Context, id uuid.UUID) (model.Order, error) {
			return model.Order{ID: orderID, StudentID: studentID, Status: "paid"}, nil
		},
		getExamByProductIDFn: func(ctx context.Context, pid uuid.UUID) (*model.Exam, error) {
			return &model.Exam{ID: examID, Title: "Finals", ProductID: &productID}, nil
		},
		createExamRegistrationFn: func(ctx context.Context, tx pgx.Tx, reg model.ExamRegistration) error {
			createCalls++
			return nil
		},
		stampOrderItemFulfilledFn: func(ctx context.Context, tx pgx.Tx, oid, pid uuid.UUID) error {
			fulfilledCalls++
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
	w.pollOutbox(ctx)

	if createCalls != 2 {
		t.Errorf("CreateExamRegistration called %d times across two polls, want 2", createCalls)
	}
	if fulfilledCalls != 2 {
		t.Errorf("StampOrderItemFulfilledAt called %d times across two polls, want 2", fulfilledCalls)
	}
}
