package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/queue"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
)

type Scheduler struct {
	storage         storage.SchedulerStorage
	publisher       queue.Publisher
	logger          logger.Logger
	scanInterval    time.Duration
	cleanupInterval time.Duration
	now             func() time.Time
}

type Config struct {
	ScanInterval    time.Duration
	CleanupInterval time.Duration
}

func New(storage storage.SchedulerStorage, publisher queue.Publisher, logger logger.Logger, config Config) *Scheduler {
	if config.ScanInterval <= 0 {
		config.ScanInterval = time.Minute
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = time.Hour
	}

	return &Scheduler{
		storage:         storage,
		publisher:       publisher,
		logger:          logger,
		scanInterval:    config.ScanInterval,
		cleanupInterval: config.CleanupInterval,
		now:             time.Now,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	if err := s.NotifyDueEvents(ctx); err != nil {
		s.logError("failed to notify due events: " + err.Error())
	}
	if err := s.DeleteOldEvents(); err != nil {
		s.logError("failed to delete old events: " + err.Error())
	}

	scanTicker := time.NewTicker(s.scanInterval)
	defer scanTicker.Stop()

	cleanupTicker := time.NewTicker(s.cleanupInterval)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-scanTicker.C:
			if err := s.NotifyDueEvents(ctx); err != nil {
				s.logError("failed to notify due events: " + err.Error())
			}
		case <-cleanupTicker.C:
			if err := s.DeleteOldEvents(); err != nil {
				s.logError("failed to delete old events: " + err.Error())
			}
		}
	}
}

func (s *Scheduler) NotifyDueEvents(ctx context.Context) error {
	now := s.now().UTC()
	events, err := s.storage.ListEventsToNotify(now)
	if err != nil {
		return fmt.Errorf("list events to notify: %w", err)
	}

	for _, event := range events {
		notification := storage.Notification{
			EventID:    event.ID.String(),
			Title:      event.Title,
			Date:       event.Date,
			UserToSend: event.UserID,
		}

		message, err := queue.MarshalJSON(notification)
		if err != nil {
			return fmt.Errorf("marshal notification for event %s: %w", event.ID, err)
		}
		if err := s.publisher.Publish(ctx, message); err != nil {
			return fmt.Errorf("publish notification for event %s: %w", event.ID, err)
		}
		if err := s.storage.MarkEventNotified(event.ID, now); err != nil {
			return fmt.Errorf("mark event %s notified: %w", event.ID, err)
		}
	}

	if len(events) > 0 {
		s.logInfo(fmt.Sprintf("published %d notifications", len(events)))
	}

	return nil
}

func (s *Scheduler) DeleteOldEvents() error {
	before := s.now().UTC().AddDate(-1, 0, 0)
	if err := s.storage.DeleteEventsBefore(before); err != nil {
		return fmt.Errorf("delete events before %s: %w", before.Format(time.RFC3339), err)
	}

	return nil
}

func (s *Scheduler) SetNow(now func() time.Time) {
	s.now = now
}

func (s *Scheduler) logInfo(msg string) {
	if s.logger != nil {
		s.logger.Info(msg)
	}
}

func (s *Scheduler) logError(msg string) {
	if s.logger != nil {
		s.logger.Error(msg)
	}
}
