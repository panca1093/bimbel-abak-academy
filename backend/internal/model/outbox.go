package model

import (
	"encoding/json"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID          int64
	AggregateID uuid.UUID
	EventType   string
	Payload     json.RawMessage
	CreatedAt   string
	Attempts    int
	LastError   *string
}
