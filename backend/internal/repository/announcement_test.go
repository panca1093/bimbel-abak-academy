package repository

import (
	"context"
	"testing"
	"time"

	"akademi-bimbel/internal/model"
)

// Compile-time check: *Repository must implement all announcement methods.
var _ interface {
	CreateAnnouncement(context.Context, *model.Announcement) error
	GetAnnouncementByID(context.Context, string) (*model.Announcement, error)
	ListAnnouncements(context.Context) ([]model.Announcement, error)
	UpdateAnnouncement(context.Context, string, *model.Announcement) error
	DeleteAnnouncement(context.Context, string) error
	ClaimDueAnnouncements(context.Context, time.Time, int) ([]model.Announcement, error)
	MarkAnnouncementSent(context.Context, string, time.Time, int) error
	ListActiveUserEmails(context.Context, string) ([]string, error)
} = (*Repository)(nil)

func TestAnnouncementMethodsExist(t *testing.T) {
	r := &Repository{}
	ctx := context.Background()

	// Function pointer checks verify method signatures at compile time.
	_ = r.CreateAnnouncement
	_ = r.GetAnnouncementByID
	_ = r.ListAnnouncements
	_ = r.UpdateAnnouncement
	_ = r.DeleteAnnouncement
	_ = r.ClaimDueAnnouncements
	_ = r.MarkAnnouncementSent
	_ = r.ListActiveUserEmails

	_ = ctx
}
