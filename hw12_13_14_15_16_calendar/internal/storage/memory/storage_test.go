package memorystorage

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
)

func TestStorageCreateAndListEvents(t *testing.T) {
	t.Parallel()

	db := New()
	userID := uuid.New()
	event := newEvent(userID, time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))

	if err := db.CreateEvent(event); err != nil {
		t.Fatalf("create event: %v", err)
	}

	events, err := db.ListEventsForDay(time.Date(2026, 6, 29, 23, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("list day events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != event.ID {
		t.Fatalf("expected event id %s, got %s", event.ID, events[0].ID)
	}
}

func TestStorageUpdateEvent(t *testing.T) {
	t.Parallel()

	db := New()
	userID := uuid.New()
	event := newEvent(userID, time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))

	if err := db.CreateEvent(event); err != nil {
		t.Fatalf("create event: %v", err)
	}

	updated := event
	updated.Title = "updated"
	updated.Date = time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC)

	if err := db.UpdateEvent(event.ID, updated); err != nil {
		t.Fatalf("update event: %v", err)
	}

	events, err := db.ListEventsForDay(updated.Date)
	if err != nil {
		t.Fatalf("list updated day events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "updated" {
		t.Fatalf("expected updated title, got %q", events[0].Title)
	}

	oldEvents, err := db.ListEventsForDay(event.Date)
	if err != nil {
		t.Fatalf("list old day events: %v", err)
	}
	if len(oldEvents) != 0 {
		t.Fatalf("expected old date to be empty, got %d events", len(oldEvents))
	}
}

func TestStorageDeleteEvent(t *testing.T) {
	t.Parallel()

	db := New()
	event := newEvent(uuid.New(), time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))

	if err := db.CreateEvent(event); err != nil {
		t.Fatalf("create event: %v", err)
	}
	if err := db.DeleteEvent(event.ID); err != nil {
		t.Fatalf("delete event: %v", err)
	}

	events, err := db.ListEventsForDay(event.Date)
	if err != nil {
		t.Fatalf("list day events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}

	if err := db.DeleteEvent(event.ID); !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestStorageDateBusy(t *testing.T) {
	t.Parallel()

	db := New()
	userID := uuid.New()
	event := newEvent(userID, time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))
	overlapping := newEvent(userID, time.Date(2026, 6, 29, 10, 30, 0, 0, time.UTC))

	if err := db.CreateEvent(event); err != nil {
		t.Fatalf("create event: %v", err)
	}
	if err := db.CreateEvent(overlapping); !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got %v", err)
	}
}

func TestStorageListsEventsForWeekAndMonth(t *testing.T) {
	t.Parallel()

	db := New()
	userID := uuid.New()
	juneEvent := newEvent(userID, time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))
	julyEvent := newEvent(userID, time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC))
	nextWeekEvent := newEvent(userID, time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC))

	for _, event := range []storage.Event{juneEvent, julyEvent, nextWeekEvent} {
		if err := db.CreateEvent(event); err != nil {
			t.Fatalf("create event %s: %v", event.ID, err)
		}
	}

	weekEvents, err := db.ListEventsForWeek(time.Date(2026, 6, 29, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("list week events: %v", err)
	}
	if len(weekEvents) != 2 {
		t.Fatalf("expected 2 week events, got %d", len(weekEvents))
	}

	monthEvents, err := db.ListEventsForMonth(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("list month events: %v", err)
	}
	if len(monthEvents) != 1 {
		t.Fatalf("expected 1 month event, got %d", len(monthEvents))
	}
	if monthEvents[0].ID != juneEvent.ID {
		t.Fatalf("expected june event id %s, got %s", juneEvent.ID, monthEvents[0].ID)
	}
}

func newEvent(userID uuid.UUID, date time.Time) storage.Event {
	return storage.Event{
		ID:       uuid.New(),
		Title:    "event",
		Date:     date,
		Duration: time.Hour,
		UserID:   userID,
	}
}
