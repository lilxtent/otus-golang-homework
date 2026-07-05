package sender

import (
	"context"
	"fmt"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/queue"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
)

type Sender struct {
	consumer        queue.Consumer
	statusPublisher queue.Publisher
	logger          logger.Logger
}

func New(consumer queue.Consumer, statusPublisher queue.Publisher, logger logger.Logger) *Sender {
	return &Sender{
		consumer:        consumer,
		statusPublisher: statusPublisher,
		logger:          logger,
	}
}

func (s *Sender) Run(ctx context.Context) error {
	return s.consumer.Consume(ctx, s.handleMessage)
}

func (s *Sender) handleMessage(ctx context.Context, message queue.Message) error {
	var notification storage.Notification
	if err := queue.UnmarshalJSON(message, &notification); err != nil {
		return fmt.Errorf("unmarshal notification: %w", err)
	}

	s.logInfo(fmt.Sprintf(
		"notification: event_id=%s title=%q date=%s user_id=%s",
		notification.EventID,
		notification.Title,
		notification.Date.Format("2006-01-02T15:04:05Z07:00"),
		notification.UserToSend,
	))

	status, err := queue.MarshalJSON(storage.NotificationStatus{
		EventID: notification.EventID,
		Status:  "sent",
		SentAt:  time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("marshal notification status: %w", err)
	}
	if err := s.statusPublisher.Publish(ctx, status); err != nil {
		return fmt.Errorf("publish notification status: %w", err)
	}

	return nil
}

func (s *Sender) logInfo(msg string) {
	if s.logger != nil {
		s.logger.Info(msg)
	}
}
