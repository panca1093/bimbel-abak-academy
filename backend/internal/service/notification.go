package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type PurchaseNotification struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	OrderID     uuid.UUID `json:"order_id"`
	StudentName string    `json:"student_name"`
	Amount      int64     `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	Read        bool      `json:"read"`
}

type NotifFilter struct {
	Type       string
	UnreadOnly bool
	Cursor     string
	Limit      int
}

func (s *Service) PushPurchaseNotification(ctx context.Context, adminRole string, notif PurchaseNotification) error {
	key := "notif:" + adminRole
	data, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	score := float64(notif.CreatedAt.UnixMilli())
	if err := s.rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: string(data)}).Err(); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListNotifications(ctx context.Context, role string, filter NotifFilter) ([]PurchaseNotification, string, error) {
	key := "notif:" + role

	// Default limit if not set
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	// Parse cursor if provided (it's the score/index of the last item)
	var offset int64 = 0
	if filter.Cursor != "" {
		parsed, err := strconv.ParseInt(filter.Cursor, 10, 64)
		if err == nil {
			offset = parsed
		}
	}

	// ZREVRANGEBYSCORE: get in reverse time order (highest scores first = newest first)
	// We fetch limit+1 to determine if there are more items
	members, err := s.rdb.ZRevRange(ctx, key, offset, offset+int64(filter.Limit)).Result()
	if err != nil {
		return nil, "", err
	}

	var notifications []PurchaseNotification
	var nextCursor string

	for i, member := range members {
		var notif PurchaseNotification
		if err := json.Unmarshal([]byte(member), &notif); err != nil {
			continue
		}

		// Apply type filter
		if filter.Type != "" && notif.Type != filter.Type {
			continue
		}

		// Check if read status matches filter
		readKey := "notif_read:" + role + ":" + notif.ID
		isRead, err := s.rdb.Exists(ctx, readKey).Result()
		if err != nil {
			continue
		}
		notif.Read = isRead == 1

		if filter.UnreadOnly && notif.Read {
			continue
		}

		notifications = append(notifications, notif)

		// Set next cursor if we're at the limit
		if len(notifications) >= filter.Limit {
			if i < len(members)-1 {
				nextCursor = strconv.FormatInt(offset+int64(i)+1, 10)
			}
			break
		}
	}

	return notifications, nextCursor, nil
}

func (s *Service) MarkNotificationRead(ctx context.Context, role, id string) error {
	key := "notif_read:" + role + ":" + id
	// SET NX: only set if key does not exist
	if err := s.rdb.SetNX(ctx, key, "true", 0).Err(); err != nil {
		return err
	}
	return nil
}
