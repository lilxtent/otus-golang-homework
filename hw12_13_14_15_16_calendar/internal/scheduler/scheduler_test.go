package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/queue"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNotifyDueEventsPublishesNotificationsAndMarksEvents(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	event := storage.Event{
		ID:     uuid.New(),
		Title:  "standup",
		Date:   now.Add(30 * time.Minute),
		UserID: uuid.New(),
	}
	strg := &fakeStorage{eventsToNotify: []storage.Event{event}}
	publisher := &fakePublisher{}
	service := New(strg, publisher, nil, Config{})
	service.SetNow(func() time.Time { return now })

	err := service.NotifyDueEvents(context.Background())
	require.NoError(t, err)
	require.Len(t, publisher.messages, 1)

	var notification storage.Notification
	err = queue.UnmarshalJSON(publisher.messages[0], &notification)
	require.NoError(t, err)
	require.Equal(t, event.ID.String(), notification.EventID)
	require.Equal(t, event.Title, notification.Title)
	require.Equal(t, event.Date, notification.Date)
	require.Equal(t, event.UserID, notification.UserToSend)

	require.Equal(t, []uuid.UUID{event.ID}, strg.markedIDs)
	require.Equal(t, []time.Time{now}, strg.markedAt)
}

func TestNotifyDueEventsDoesNotMarkEventWhenPublishFails(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	event := storage.Event{ID: uuid.New(), Title: "standup", Date: now.Add(time.Hour), UserID: uuid.New()}
	strg := &fakeStorage{eventsToNotify: []storage.Event{event}}
	publisher := &fakePublisher{err: errors.New("publish failed")}
	service := New(strg, publisher, nil, Config{})
	service.SetNow(func() time.Time { return now })

	err := service.NotifyDueEvents(context.Background())
	require.ErrorContains(t, err, "publish notification")
	require.Empty(t, strg.markedIDs)
}

func TestDeleteOldEventsUsesOneYearThreshold(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	strg := &fakeStorage{}
	service := New(strg, &fakePublisher{}, nil, Config{})
	service.SetNow(func() time.Time { return now })

	err := service.DeleteOldEvents()
	require.NoError(t, err)
	require.Equal(t, now.AddDate(-1, 0, 0), strg.deletedBefore)
}

type fakeStorage struct {
	eventsToNotify []storage.Event
	markedIDs      []uuid.UUID
	markedAt       []time.Time
	deletedBefore  time.Time
}

func (s *fakeStorage) ListEventsToNotify(time.Time) ([]storage.Event, error) {
	return s.eventsToNotify, nil
}

func (s *fakeStorage) MarkEventNotified(id uuid.UUID, notifiedAt time.Time) error {
	s.markedIDs = append(s.markedIDs, id)
	s.markedAt = append(s.markedAt, notifiedAt)
	return nil
}

func (s *fakeStorage) DeleteEventsBefore(before time.Time) error {
	s.deletedBefore = before
	return nil
}

type fakePublisher struct {
	messages []queue.Message
	err      error
}

func (p *fakePublisher) Publish(_ context.Context, message queue.Message) error {
	if p.err != nil {
		return p.err
	}

	p.messages = append(p.messages, message)
	return nil
}
