package sender

import (
	"context"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/queue"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRunConsumesAndLogsNotification(t *testing.T) {
	t.Parallel()

	notification := storage.Notification{
		EventID:    uuid.New().String(),
		Title:      "standup",
		Date:       time.Date(2026, 7, 2, 12, 30, 0, 0, time.UTC),
		UserToSend: uuid.New(),
	}
	message, err := queue.MarshalJSON(notification)
	require.NoError(t, err)

	consumer := &fakeConsumer{message: message}
	logger := &fakeLogger{}
	service := New(consumer, logger)

	err = service.Run(context.Background())
	require.NoError(t, err)
	require.Len(t, logger.info, 1)
	require.Contains(t, logger.info[0], notification.EventID)
	require.Contains(t, logger.info[0], notification.Title)
	require.Contains(t, logger.info[0], notification.UserToSend.String())
}

func TestRunReturnsErrorForInvalidNotification(t *testing.T) {
	t.Parallel()

	consumer := &fakeConsumer{message: queue.Message{Body: []byte("{")}}
	service := New(consumer, &fakeLogger{})

	err := service.Run(context.Background())
	require.ErrorContains(t, err, "unmarshal notification")
}

type fakeConsumer struct {
	message queue.Message
	err     error
}

func (c *fakeConsumer) Consume(ctx context.Context, handler queue.Handler) error {
	if c.err != nil {
		return c.err
	}

	if err := handler(ctx, c.message); err != nil {
		return err
	}

	return nil
}

type fakeLogger struct {
	info  []string
	error []string
}

func (l *fakeLogger) Info(msg string) {
	l.info = append(l.info, msg)
}

func (l *fakeLogger) Error(msg string) {
	l.error = append(l.error, msg)
}
