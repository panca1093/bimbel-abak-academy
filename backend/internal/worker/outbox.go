package worker

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"akademi-bimbel/internal/model"
)

// OrderPaidPayload unmarshals from the outbox event payload written by service.OrderPaidPayload
type OrderPaidPayload struct {
	OrderID uuid.UUID      `json:"order_id"`
	Items   []OrderItemMini `json:"items"`
}

// OrderItemMini contains minimal item info for access provisioning
type OrderItemMini struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductType string    `json:"product_type"`
}

// outboxRepository interface defines the methods needed from the repository
type outboxRepository interface {
	ClaimOutboxEvents(context.Context, int) ([]model.OutboxEvent, error)
	MarkOutboxProcessed(context.Context, pgx.Tx, int64) error
	GetOrderByID(context.Context, uuid.UUID) (model.Order, error)
	SetOrderStatus(context.Context, pgx.Tx, uuid.UUID, string, string) error
	CreateCourseSession(context.Context, pgx.Tx, model.CourseSession) error
	GetCoursesByProductID(context.Context, uuid.UUID) ([]model.Course, error)
	BeginTx(context.Context) (pgx.Tx, error)
	GetExpiredPaymentOrders(context.Context, int) ([]uuid.UUID, error)
	GetExamByProductID(context.Context, uuid.UUID) (*model.Exam, error)
	CreateExamRegistration(context.Context, pgx.Tx, model.ExamRegistration) error
	StampOrderItemFulfilledAt(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) error
}

type Worker struct {
	pool                     *pgxpool.Pool
	rdb                      *redis.Client
	repo                     outboxRepository
	interval                 time.Duration
	sweeperInterval          time.Duration
	announcementPollInterval time.Duration
	dispatcher               announcementDispatcher
	jobRepo                  jobRepository
	objectStore              objectStore
	svc                      studentBulkProcessor
	jobPollInterval          time.Duration
	privateBucket            string
}

func New(pool *pgxpool.Pool, rdb *redis.Client, repo outboxRepository, interval, sweeperInterval, announcementPollInterval time.Duration, dispatcher announcementDispatcher, jobRepo jobRepository, objectStore objectStore, svc studentBulkProcessor, jobPollInterval time.Duration, privateBucket string) *Worker {
	return &Worker{
		pool:                     pool,
		rdb:                      rdb,
		repo:                     repo,
		interval:                 interval,
		sweeperInterval:          sweeperInterval,
		announcementPollInterval: announcementPollInterval,
		dispatcher:               dispatcher,
		jobRepo:                  jobRepo,
		objectStore:              objectStore,
		svc:                      svc,
		jobPollInterval:          jobPollInterval,
		privateBucket:            privateBucket,
	}
}

// Run polls the transactional outbox, runs the stale-payment sweeper, dispatches
// due announcements, and polls the job table until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	outboxTicker := time.NewTicker(w.interval)
	defer outboxTicker.Stop()

	sweeperTicker := time.NewTicker(w.sweeperInterval)
	defer sweeperTicker.Stop()

	announcementTicker := time.NewTicker(w.announcementPollInterval)
	defer announcementTicker.Stop()

	jobTicker := time.NewTicker(w.jobPollInterval)
	defer jobTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sweeperTicker.C:
				w.sweepStalePayments(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-announcementTicker.C:
				w.pollAnnouncements(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-jobTicker.C:
				w.pollJobs(ctx)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-outboxTicker.C:
			w.pollOutbox(ctx)
		}
	}
}

func (w *Worker) pollOutbox(ctx context.Context) {
	events, err := w.repo.ClaimOutboxEvents(ctx, 10)
	if err != nil {
		slog.Error("claim outbox events", "err", err)
		return
	}

	for _, event := range events {
		switch event.EventType {
		case "OrderPaid":
			w.handleOrderPaid(ctx, event)
		default:
			slog.Warn("unknown event type", "type", event.EventType)
		}
	}
}

func (w *Worker) handleOrderPaid(ctx context.Context, event model.OutboxEvent) {
	var payload OrderPaidPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		slog.Error("unmarshal OrderPaid payload", "event_id", event.ID, "err", err)
		return
	}

	order, err := w.repo.GetOrderByID(ctx, payload.OrderID)
	if err != nil {
		slog.Error("get order by id", "order_id", payload.OrderID, "err", err)
		return
	}

	tx, err := w.repo.BeginTx(ctx)
	if err != nil {
		slog.Error("begin tx", "err", err)
		return
	}
	defer tx.Rollback(ctx)

	// Provision access for each item; track whether any physical item requires shipping.
	hasPhysicalItem := false
	for _, item := range payload.Items {
		switch item.ProductType {
		case "course":
			courses, err := w.repo.GetCoursesByProductID(ctx, item.ProductID)
			if err != nil {
				slog.Error("get courses by product id", "order_id", payload.OrderID, "product_id", item.ProductID, "err", err)
				return
			}
			if len(courses) == 0 {
				slog.Warn("no courses linked to product, skipping", "order_id", payload.OrderID, "product_id", item.ProductID)
				continue
			}
			for _, course := range courses {
				session := model.CourseSession{
					StudentID:  order.StudentID,
					CourseID:   course.ID,
					OrderID:    &payload.OrderID,
					Status:     "active",
					Source:     "order",
					EnrolledAt: time.Now(),
				}
				if err := w.repo.CreateCourseSession(ctx, tx, session); err != nil {
					slog.Error("create course session", "order_id", payload.OrderID, "course_id", course.ID, "err", err)
					return
				}
			}
		case "exam":
			exam, err := w.repo.GetExamByProductID(ctx, item.ProductID)
			if err != nil {
				slog.Error("get exam by product id", "order_id", payload.OrderID, "product_id", item.ProductID, "err", err)
				return
			}
			if err := w.repo.CreateExamRegistration(ctx, tx, model.ExamRegistration{
				StudentID: order.StudentID,
				ExamID:    exam.ID,
				Token:     generateToken(),
				Status:    "registered",
			}); err != nil {
				slog.Error("create exam registration", "order_id", payload.OrderID, "err", err)
				return
			}
			if err := w.repo.StampOrderItemFulfilledAt(ctx, tx, payload.OrderID, item.ProductID); err != nil {
				slog.Error("stamp order_item fulfilled_at", "order_id", payload.OrderID, "product_id", item.ProductID, "err", err)
				return
			}
		case "book":
			hasPhysicalItem = true
		default:
			slog.Warn("unknown product type in order", "product_type", item.ProductType, "product_id", item.ProductID)
		}
	}

	// Digital-only orders complete immediately; physical orders need admin to ship.
	nextStatus := "completed"
	if hasPhysicalItem {
		nextStatus = "processing"
	}
	if err := w.repo.SetOrderStatus(ctx, tx, payload.OrderID, nextStatus, ""); err != nil {
		slog.Error("set order status", "order_id", payload.OrderID, "err", err)
		return
	}

	// Mark outbox event processed
	if err := w.repo.MarkOutboxProcessed(ctx, tx, event.ID); err != nil {
		slog.Error("mark outbox processed", "event_id", event.ID, "err", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("commit tx", "err", err)
		return
	}

	slog.Info("order paid processed", "order_id", payload.OrderID, "event_id", event.ID)
}

func (w *Worker) sweepStalePayments(ctx context.Context) {
	orderIDs, err := w.repo.GetExpiredPaymentOrders(ctx, 50)
	if err != nil {
		slog.Error("get expired payment orders", "err", err)
		return
	}

	for _, orderID := range orderIDs {
		tx, err := w.repo.BeginTx(ctx)
		if err != nil {
			slog.Error("begin tx for sweeper", "err", err)
			continue
		}

		// Re-check status inside transaction
		currentOrder, err := w.repo.GetOrderByID(ctx, orderID)
		if err != nil {
			slog.Error("get order for recheck", "order_id", orderID, "err", err)
			tx.Rollback(ctx)
			continue
		}

		if currentOrder.Status != "payment_pending" {
			tx.Rollback(ctx)
			continue
		}

		// Set order status to payment_expired
		if err := w.repo.SetOrderStatus(ctx, tx, orderID, "payment_expired", ""); err != nil {
			slog.Error("set payment expired status", "order_id", orderID, "err", err)
			tx.Rollback(ctx)
			continue
		}

		// Restore stock for each item
		for _, item := range currentOrder.Items {
			_, err := tx.Exec(ctx,
				`UPDATE product SET stock = stock + $1, updated_at = now() WHERE id = $2`,
				item.Qty, item.ProductID,
			)
			if err != nil {
				slog.Error("restore stock", "order_id", orderID, "product_id", item.ProductID, "err", err)
				tx.Rollback(ctx)
				continue
			}
		}

		if err := tx.Commit(ctx); err != nil {
			slog.Error("commit sweeper tx", "order_id", orderID, "err", err)
			continue
		}

		slog.Info("stale payment expired", "order_id", orderID)
	}
}

func generateToken() string {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read never fails on supported platforms; a constant
		// fallback would be a guessable check-in credential.
		panic(err)
	}
	return strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)[:8])
}
