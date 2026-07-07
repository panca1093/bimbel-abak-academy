package worker

import (
	"context"
	"log/slog"
)

// announcementDispatcher defines the service method the announcement ticker calls.
type announcementDispatcher interface {
	DispatchDueAnnouncements(ctx context.Context, limit int) (int, error)
}

// pollAnnouncements calls the dispatcher and logs the result.
// Safe to call when dispatcher is nil (noop — no announcements dispatched).
func (w *Worker) pollAnnouncements(ctx context.Context) {
	if w.dispatcher == nil {
		return
	}
	count, err := w.dispatcher.DispatchDueAnnouncements(ctx, 50)
	if err != nil {
		slog.Error("dispatch due announcements", "err", err)
		return
	}
	slog.Info("dispatched due announcements", "count", count)
}
