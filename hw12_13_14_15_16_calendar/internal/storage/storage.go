package storage

import (
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	CreateEvent(event Event) (Event, error)
	UpdateEvent(id uuid.UUID, event Event) error
	DeleteEvent(id uuid.UUID) error
	ListEventsForDay(date time.Time) ([]Event, error)
	ListEventsForWeek(startOfWeek time.Time) ([]Event, error)
	ListEventsForMonth(startOfMonth time.Time) ([]Event, error)
}
