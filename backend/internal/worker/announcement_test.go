package worker

import (
	"context"
	"errors"
	"testing"
)

type mockAnnouncementDispatcher struct {
	dispatchDueAnnouncementsFn func(context.Context, int) (int, error)
}

func (m *mockAnnouncementDispatcher) DispatchDueAnnouncements(ctx context.Context, limit int) (int, error) {
	return m.dispatchDueAnnouncementsFn(ctx, limit)
}

func TestPollAnnouncementsCallsDispatcher(t *testing.T) {
	ctx := context.Background()
	var called bool

	dispatcher := &mockAnnouncementDispatcher{
		dispatchDueAnnouncementsFn: func(ctx context.Context, limit int) (int, error) {
			called = true
			if limit != 50 {
				t.Errorf("expected limit 50, got %d", limit)
			}
			return 3, nil
		},
	}

	w := &Worker{dispatcher: dispatcher}
	w.pollAnnouncements(ctx)

	if !called {
		t.Error("expected DispatchDueAnnouncements to be called")
	}
}

func TestPollAnnouncementsDoesNotPanicOnError(t *testing.T) {
	ctx := context.Background()

	dispatcher := &mockAnnouncementDispatcher{
		dispatchDueAnnouncementsFn: func(ctx context.Context, limit int) (int, error) {
			return 0, errors.New("dispatch failed")
		},
	}

	w := &Worker{dispatcher: dispatcher}
	w.pollAnnouncements(ctx)

	// If we got here without panicking, the test passes.
}
