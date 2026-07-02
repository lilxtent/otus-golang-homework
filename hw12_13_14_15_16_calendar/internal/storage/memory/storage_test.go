package memorystorage

import (
	"errors"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStorageCreateAndListEvents(t *testing.T) {
	t.Parallel()

	db := New()
	userID := uuid.New()
	event := newEvent(userID, time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))

	created, err := db.CreateEvent(event)
	require.NoError(t, err)
	require.Equal(t, event.ID, created.ID)

	events, err := db.ListEventsForDay(time.Date(2026, 6, 29, 23, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, event.ID, events[0].ID)
}

func TestStorageCreateEventGeneratesID(t *testing.T) {
	t.Parallel()

	db := New()
	event := newEvent(uuid.New(), time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))
	event.ID = uuid.Nil

	created, err := db.CreateEvent(event)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)
}

func TestStorageUpdateEvent(t *testing.T) {
	t.Parallel()

	db := New()
	userID := uuid.New()
	event := newEvent(userID, time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))

	_, err := db.CreateEvent(event)
	require.NoError(t, err)

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

	_, err := db.CreateEvent(event)
	require.NoError(t, err)
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

	_, err := db.CreateEvent(event)
	require.NoError(t, err)
	_, err = db.CreateEvent(overlapping)
	require.ErrorIs(t, err, storage.ErrDateBusy)
}

func TestStorageListsEventsForWeekAndMonth(t *testing.T) {
	t.Parallel()

	db := New()
	userID := uuid.New()
	juneEvent := newEvent(userID, time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC))
	julyEvent := newEvent(userID, time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC))
	nextWeekEvent := newEvent(userID, time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC))

	for _, event := range []storage.Event{juneEvent, julyEvent, nextWeekEvent} {
		_, err := db.CreateEvent(event)
		require.NoError(t, err)
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

func TestStorageListsEventsToNotify(t *testing.T) {
	t.Parallel()

	db := New()
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	notifyBefore := time.Hour
	notifiedAt := now.Add(-time.Minute)

	due := newEvent(uuid.New(), now.Add(30*time.Minute))
	due.NotifyBefore = &notifyBefore

	future := newEvent(uuid.New(), now.Add(2*time.Hour))
	future.NotifyBefore = &notifyBefore

	withoutNotification := newEvent(uuid.New(), now.Add(30*time.Minute))

	past := newEvent(uuid.New(), now.Add(-time.Minute))
	past.NotifyBefore = &notifyBefore

	alreadyNotified := newEvent(uuid.New(), now.Add(30*time.Minute))
	alreadyNotified.NotifyBefore = &notifyBefore
	alreadyNotified.NotifiedAt = &notifiedAt

	for _, event := range []storage.Event{due, future, withoutNotification, past, alreadyNotified} {
		_, err := db.CreateEvent(event)
		require.NoError(t, err)
	}

	events, err := db.ListEventsToNotify(now)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, due.ID, events[0].ID)
}

func TestStorageMarkEventNotified(t *testing.T) {
	t.Parallel()

	db := New()
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	notifyBefore := time.Hour
	event := newEvent(uuid.New(), now.Add(30*time.Minute))
	event.NotifyBefore = &notifyBefore

	_, err := db.CreateEvent(event)
	require.NoError(t, err)

	err = db.MarkEventNotified(event.ID, now)
	require.NoError(t, err)

	events, err := db.ListEventsToNotify(now)
	require.NoError(t, err)
	require.Empty(t, events)

	events, err = db.ListEventsForDay(event.Date)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.NotNil(t, events[0].NotifiedAt)
	require.Equal(t, now, *events[0].NotifiedAt)
}

func TestStorageDeleteEventsBefore(t *testing.T) {
	t.Parallel()

	db := New()
	userID := uuid.New()
	oldEvent := newEvent(userID, time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC))
	borderEvent := newEvent(userID, time.Date(2025, 7, 2, 10, 0, 0, 0, time.UTC))
	freshEvent := newEvent(userID, time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC))

	for _, event := range []storage.Event{oldEvent, borderEvent, freshEvent} {
		_, err := db.CreateEvent(event)
		require.NoError(t, err)
	}

	err := db.DeleteEventsBefore(borderEvent.Date)
	require.NoError(t, err)

	oldEvents, err := db.ListEventsForDay(oldEvent.Date)
	require.NoError(t, err)
	require.Empty(t, oldEvents)

	borderEvents, err := db.ListEventsForDay(borderEvent.Date)
	require.NoError(t, err)
	require.Len(t, borderEvents, 1)

	freshEvents, err := db.ListEventsForDay(freshEvent.Date)
	require.NoError(t, err)
	require.Len(t, freshEvents, 1)
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
