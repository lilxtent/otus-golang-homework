package memorystorage

import (
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
)

const (
	eventsTable = "events"
	idIndex     = "id"
	dateIndex   = "date"
	dateKeyFmt  = "20060102150405.000000000"
)

type MemoryStorage struct {
	db *memdb.MemDB
}

type eventRecord struct {
	ID      string
	DateKey string
	Event   storage.Event
}

func New() *MemoryStorage {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			eventsTable: {
				Name: eventsTable,
				Indexes: map[string]*memdb.IndexSchema{
					idIndex: {
						Name:    idIndex,
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					dateIndex: {
						Name:    dateIndex,
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "DateKey"},
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return &MemoryStorage{db: db}
}

func (s *MemoryStorage) CreateEvent(event storage.Event) (storage.Event, error) {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	txn := s.db.Txn(true)
	defer txn.Abort()

	found, err := txn.First(eventsTable, idIndex, event.ID.String())
	if err != nil {
		return storage.Event{}, err
	}
	if found != nil {
		return storage.Event{}, storage.ErrEventAlreadyExists
	}

	if err := ensureDateAvailable(txn, event, uuid.Nil); err != nil {
		return storage.Event{}, err
	}

	if err := txn.Insert(eventsTable, newEventRecord(event)); err != nil {
		return storage.Event{}, err
	}

	txn.Commit()
	return cloneEvent(event), nil
}

func (s *MemoryStorage) UpdateEvent(id uuid.UUID, event storage.Event) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	existing, err := txn.First(eventsTable, idIndex, id.String())
	if err != nil {
		return err
	}
	if existing == nil {
		return storage.ErrEventNotFound
	}

	event.ID = id
	if err := ensureDateAvailable(txn, event, id); err != nil {
		return err
	}

	if err := txn.Delete(eventsTable, existing); err != nil {
		return err
	}
	if err := txn.Insert(eventsTable, newEventRecord(event)); err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *MemoryStorage) DeleteEvent(id uuid.UUID) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	existing, err := txn.First(eventsTable, idIndex, id.String())
	if err != nil {
		return err
	}
	if existing == nil {
		return storage.ErrEventNotFound
	}

	if err := txn.Delete(eventsTable, existing); err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *MemoryStorage) ListEventsForDay(date time.Time) ([]storage.Event, error) {
	start := dayStart(date)
	return s.listEventsBetween(start, start.AddDate(0, 0, 1))
}

func (s *MemoryStorage) ListEventsForWeek(startOfWeek time.Time) ([]storage.Event, error) {
	start := dayStart(startOfWeek)
	return s.listEventsBetween(start, start.AddDate(0, 0, 7))
}

func (s *MemoryStorage) ListEventsForMonth(startOfMonth time.Time) ([]storage.Event, error) {
	start := dayStart(startOfMonth)
	return s.listEventsBetween(start, start.AddDate(0, 1, 0))
}

func (s *MemoryStorage) listEventsBetween(start, end time.Time) ([]storage.Event, error) {
	txn := s.db.Txn(false)
	defer txn.Abort()

	it, err := txn.LowerBound(eventsTable, dateIndex, dateKey(start))
	if err != nil {
		return nil, err
	}

	events := make([]storage.Event, 0)
	endKey := dateKey(end)
	for raw := it.Next(); raw != nil; raw = it.Next() {
		record := raw.(*eventRecord)
		if record.DateKey >= endKey {
			break
		}
		events = append(events, cloneEvent(record.Event))
	}

	return events, nil
}

func ensureDateAvailable(txn *memdb.Txn, event storage.Event, excludedID uuid.UUID) error {
	it, err := txn.LowerBound(eventsTable, idIndex, "")
	if err != nil {
		return err
	}

	for raw := it.Next(); raw != nil; raw = it.Next() {
		record := raw.(*eventRecord)
		stored := record.Event
		if stored.ID == excludedID || stored.UserID != event.UserID {
			continue
		}
		if eventsOverlap(stored, event) {
			return storage.ErrDateBusy
		}
	}

	return nil
}

func eventsOverlap(left, right storage.Event) bool {
	leftStart := left.Date
	leftEnd := leftStart.Add(left.Duration)
	rightStart := right.Date
	rightEnd := rightStart.Add(right.Duration)

	if !leftEnd.After(leftStart) {
		leftEnd = leftStart
	}
	if !rightEnd.After(rightStart) {
		rightEnd = rightStart
	}

	return leftStart.Before(rightEnd) && rightStart.Before(leftEnd)
}

func newEventRecord(event storage.Event) *eventRecord {
	event = cloneEvent(event)

	return &eventRecord{
		ID:      event.ID.String(),
		DateKey: dateKey(event.Date),
		Event:   event,
	}
}

func dateKey(date time.Time) string {
	return date.UTC().Format(dateKeyFmt)
}

func dayStart(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func cloneEvent(event storage.Event) storage.Event {
	if event.Description != nil {
		description := *event.Description
		event.Description = &description
	}
	if event.NotifyBefore != nil {
		notifyBefore := *event.NotifyBefore
		event.NotifyBefore = &notifyBefore
	}

	return event
}
