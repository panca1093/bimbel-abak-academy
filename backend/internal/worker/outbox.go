package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Worker struct {
	pool     *pgxpool.Pool
	interval time.Duration
}

func New(pool *pgxpool.Pool, interval time.Duration) *Worker {
	return &Worker{pool: pool, interval: interval}
}

// Run polls the transactional outbox until ctx is cancelled. Event dispatch
// (OrderPaid -> access provisioning, PDF generation, notifications) lands here
// as feature work; the skeleton only proves the poll loop and DB connection.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *Worker) poll(ctx context.Context) {
	var pending int
	err := w.pool.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE processed_at IS NULL`).Scan(&pending)
	if err != nil {
		slog.Error("outbox poll", "err", err)
		return
	}
	slog.Info("outbox poll", "pending", pending)
}
