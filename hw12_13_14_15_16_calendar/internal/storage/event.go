package storage

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID           uuid.UUID
	Title        string
	Date         time.Time
	Duration     time.Duration
	Description  *string
	UserID       uuid.UUID
	NotifyBefore *time.Duration
}

type Notification struct {
	EventID    string
	Title      string
	Date       time.Time
	UserToSend uuid.UUID
}
