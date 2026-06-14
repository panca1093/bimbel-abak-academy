package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestPushPurchaseNotification(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis setup failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}
	ctx := context.Background()

	orderId := uuid.New()
	notif := PurchaseNotification{
		ID:        uuid.New().String(),
		Type:      "order_confirmed",
		OrderID:   orderId,
		StudentName: "John Doe",
		Amount:    100000,
		CreatedAt: time.Now(),
		Read:      false,
	}

	err = svc.PushPurchaseNotification(ctx, RoleAdminStore, notif)
	if err != nil {
		t.Fatalf("PushPurchaseNotification failed: %v", err)
	}

	// Verify the notification was added to the sorted set
	key := "notif:" + RoleAdminStore
	members, err := rdb.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		t.Fatalf("failed to get sorted set members: %v", err)
	}

	if len(members) != 1 {
		t.Fatalf("expected 1 member in sorted set, got %d", len(members))
	}

	// Verify the member is the JSON-encoded notification
	var retrieved PurchaseNotification
	if err := json.Unmarshal([]byte(members[0]), &retrieved); err != nil {
		t.Fatalf("failed to unmarshal notification: %v", err)
	}

	if retrieved.ID != notif.ID {
		t.Errorf("notification ID mismatch: expected %s, got %s", notif.ID, retrieved.ID)
	}
	if retrieved.StudentName != notif.StudentName {
		t.Errorf("student name mismatch: expected %s, got %s", notif.StudentName, retrieved.StudentName)
	}
}

func TestListNotifications(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis setup failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}
	ctx := context.Background()

	// Push multiple notifications with different times
	role := RoleAdminStore
	key := "notif:" + role

	now := time.Now()
	notif1 := PurchaseNotification{
		ID:        uuid.New().String(),
		Type:      "order_confirmed",
		OrderID:   uuid.New(),
		StudentName: "John Doe",
		Amount:    100000,
		CreatedAt: now.Add(-2 * time.Second),
		Read:      false,
	}

	notif2 := PurchaseNotification{
		ID:        uuid.New().String(),
		Type:      "order_paid",
		OrderID:   uuid.New(),
		StudentName: "Jane Smith",
		Amount:    200000,
		CreatedAt: now.Add(-1 * time.Second),
		Read:      false,
	}

	notif3 := PurchaseNotification{
		ID:        uuid.New().String(),
		Type:      "order_confirmed",
		OrderID:   uuid.New(),
		StudentName: "Bob Jones",
		Amount:    150000,
		CreatedAt: now,
		Read:      false,
	}

	// Push notifications with calculated scores
	for _, n := range []PurchaseNotification{notif1, notif2, notif3} {
		data, _ := json.Marshal(n)
		score := float64(n.CreatedAt.UnixMilli())
		if err := rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)}).Err(); err != nil {
			t.Fatalf("failed to add notification: %v", err)
		}
	}

	// Test listing notifications in reverse time order (most recent first)
	filter := NotifFilter{
		Limit: 10,
	}

	notifications, nextCursor, err := svc.ListNotifications(ctx, role, filter)
	if err != nil {
		t.Fatalf("ListNotifications failed: %v", err)
	}

	if len(notifications) != 3 {
		t.Fatalf("expected 3 notifications, got %d", len(notifications))
	}

	// Verify reverse time order (newest first)
	if notifications[0].ID != notif3.ID {
		t.Errorf("first notification should be notif3, got %s", notifications[0].ID)
	}
	if notifications[1].ID != notif2.ID {
		t.Errorf("second notification should be notif2, got %s", notifications[1].ID)
	}
	if notifications[2].ID != notif1.ID {
		t.Errorf("third notification should be notif1, got %s", notifications[2].ID)
	}

	if nextCursor != "" {
		t.Errorf("expected empty next cursor, got %s", nextCursor)
	}
}

func TestListNotificationsWithTypeFilter(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis setup failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}
	ctx := context.Background()

	role := RoleAdminStore
	key := "notif:" + role

	now := time.Now()
	notif1 := PurchaseNotification{
		ID:          uuid.New().String(),
		Type:        "order_confirmed",
		OrderID:     uuid.New(),
		StudentName: "John Doe",
		Amount:      100000,
		CreatedAt:   now.Add(-1 * time.Second),
		Read:        false,
	}

	notif2 := PurchaseNotification{
		ID:          uuid.New().String(),
		Type:        "order_paid",
		OrderID:     uuid.New(),
		StudentName: "Jane Smith",
		Amount:      200000,
		CreatedAt:   now,
		Read:        false,
	}

	for _, n := range []PurchaseNotification{notif1, notif2} {
		data, _ := json.Marshal(n)
		score := float64(n.CreatedAt.UnixMilli())
		if err := rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)}).Err(); err != nil {
			t.Fatalf("failed to add notification: %v", err)
		}
	}

	filter := NotifFilter{
		Type:  "order_confirmed",
		Limit: 10,
	}

	notifications, _, err := svc.ListNotifications(ctx, role, filter)
	if err != nil {
		t.Fatalf("ListNotifications failed: %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifications))
	}
	if notifications[0].Type != "order_confirmed" {
		t.Errorf("expected type order_confirmed, got %s", notifications[0].Type)
	}
}

func TestMarkNotificationRead(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis setup failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}
	ctx := context.Background()

	notifId := uuid.New().String()
	role := RoleAdminStore

	// Mark notification as read
	err = svc.MarkNotificationRead(ctx, role, notifId)
	if err != nil {
		t.Fatalf("MarkNotificationRead failed: %v", err)
	}

	// Verify the read key was set
	readKey := "notif_read:" + role + ":" + notifId
	exists, err := rdb.Exists(ctx, readKey).Result()
	if err != nil {
		t.Fatalf("failed to check read key: %v", err)
	}
	if exists != 1 {
		t.Errorf("expected read key to exist, got %d", exists)
	}
}

func TestMarkNotificationReadIdempotent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis setup failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}
	ctx := context.Background()

	notifId := uuid.New().String()
	role := RoleAdminStore

	// Mark as read first time
	err = svc.MarkNotificationRead(ctx, role, notifId)
	if err != nil {
		t.Fatalf("first MarkNotificationRead failed: %v", err)
	}

	// Mark as read second time (should not error, SET NX returns false)
	err = svc.MarkNotificationRead(ctx, role, notifId)
	if err != nil {
		t.Fatalf("second MarkNotificationRead failed: %v", err)
	}
}

func TestListNotificationsUnreadOnlyFilter(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis setup failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}
	ctx := context.Background()

	role := RoleAdminStore
	key := "notif:" + role

	now := time.Now()
	notif1 := PurchaseNotification{
		ID:          uuid.New().String(),
		Type:        "order_confirmed",
		OrderID:     uuid.New(),
		StudentName: "John Doe",
		Amount:      100000,
		CreatedAt:   now.Add(-1 * time.Second),
		Read:        false,
	}

	notif2 := PurchaseNotification{
		ID:          uuid.New().String(),
		Type:        "order_paid",
		OrderID:     uuid.New(),
		StudentName: "Jane Smith",
		Amount:      200000,
		CreatedAt:   now,
		Read:        false,
	}

	for _, n := range []PurchaseNotification{notif1, notif2} {
		data, _ := json.Marshal(n)
		score := float64(n.CreatedAt.UnixMilli())
		if err := rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)}).Err(); err != nil {
			t.Fatalf("failed to add notification: %v", err)
		}
	}

	// Mark notif1 as read
	if err := svc.MarkNotificationRead(ctx, role, notif1.ID); err != nil {
		t.Fatalf("MarkNotificationRead failed: %v", err)
	}

	// List with unreadOnly=true
	filter := NotifFilter{
		UnreadOnly: true,
		Limit:      10,
	}

	notifications, _, err := svc.ListNotifications(ctx, role, filter)
	if err != nil {
		t.Fatalf("ListNotifications failed: %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf("expected 1 unread notification, got %d", len(notifications))
	}
	if notifications[0].ID != notif2.ID {
		t.Errorf("expected notif2 (unread), got %s", notifications[0].ID)
	}
}

func TestListNotificationsWithPagination(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis setup failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}
	ctx := context.Background()

	role := RoleAdminStore
	key := "notif:" + role

	// Create 5 notifications
	now := time.Now()
	notifs := make([]PurchaseNotification, 5)
	for i := 0; i < 5; i++ {
		notifs[i] = PurchaseNotification{
			ID:          uuid.New().String(),
			Type:        "order_confirmed",
			OrderID:     uuid.New(),
			StudentName: "Student " + string(rune('0'+i)),
			Amount:      100000 * int64(i+1),
			CreatedAt:   now.Add(time.Duration(-i) * time.Second),
			Read:        false,
		}
		data, _ := json.Marshal(notifs[i])
		score := float64(notifs[i].CreatedAt.UnixMilli())
		if err := rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)}).Err(); err != nil {
			t.Fatalf("failed to add notification: %v", err)
		}
	}

	// List first 2 items
	filter := NotifFilter{Limit: 2}
	page1, cursor1, err := svc.ListNotifications(ctx, role, filter)
	if err != nil {
		t.Fatalf("ListNotifications page 1 failed: %v", err)
	}

	if len(page1) != 2 {
		t.Fatalf("expected 2 notifications on page 1, got %d", len(page1))
	}
	// Should be in reverse time order (newest first = notifs[0] then notifs[1])
	if page1[0].ID != notifs[0].ID {
		t.Errorf("expected notifs[0] first, got %s", page1[0].ID)
	}

	if cursor1 == "" {
		t.Fatalf("expected non-empty cursor for page 2")
	}

	// List next 2 items with cursor
	filter2 := NotifFilter{Limit: 2, Cursor: cursor1}
	page2, _, err := svc.ListNotifications(ctx, role, filter2)
	if err != nil {
		t.Fatalf("ListNotifications page 2 failed: %v", err)
	}

	if len(page2) != 2 {
		t.Fatalf("expected 2 notifications on page 2, got %d", len(page2))
	}
}
