package model

import "time"

type Announcement struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Type           string     `json:"type"`
	Recipients     string     `json:"recipients"`
	Status         string     `json:"status"`
	ScheduledAt    *time.Time `json:"scheduled_at"`
	SentAt         *time.Time `json:"sent_at"`
	RecipientCount *int       `json:"recipient_count"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
