package internalgrpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	eventv1 "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/grpc/pb/event/v1"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCreateAndListEvents(t *testing.T) {
	t.Parallel()

	service := newTestEventService()
	userID := uuid.New()

	created, err := service.CreateEvent(context.Background(), &eventv1.CreateEventRequest{
		Event: newTestEvent(userID, "demo", time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.GetEvent().GetId())

	response, err := service.ListEventsForDay(context.Background(), &eventv1.ListEventsRequest{
		Date: timestamppb.New(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
	require.Len(t, response.GetEvents(), 1)
	require.Equal(t, created.GetEvent().GetId(), response.GetEvents()[0].GetId())
	require.Equal(t, "demo", response.GetEvents()[0].GetTitle())
	require.Equal(t, "calendar api", response.GetEvents()[0].GetDescription().GetValue())
}

func TestUpdateAndDeleteEvent(t *testing.T) {
	t.Parallel()

	service := newTestEventService()
	userID := uuid.New()

	created, err := service.CreateEvent(context.Background(), &eventv1.CreateEventRequest{
		Event: newTestEvent(userID, "demo", time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)

	updated, err := service.UpdateEvent(context.Background(), &eventv1.UpdateEventRequest{
		Id:    created.GetEvent().GetId(),
		Event: newTestEvent(userID, "updated", time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
	require.Equal(t, created.GetEvent().GetId(), updated.GetEvent().GetId())
	require.Equal(t, "updated", updated.GetEvent().GetTitle())

	_, err = service.DeleteEvent(context.Background(), &eventv1.DeleteEventRequest{Id: created.GetEvent().GetId()})
	require.NoError(t, err)

	_, err = service.DeleteEvent(context.Background(), &eventv1.DeleteEventRequest{Id: created.GetEvent().GetId()})
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestCreateEventValidation(t *testing.T) {
	t.Parallel()

	service := newTestEventService()

	_, err := service.CreateEvent(context.Background(), &eventv1.CreateEventRequest{
		Event: &eventv1.Event{Title: "bad"},
	})

	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateEventDateBusy(t *testing.T) {
	t.Parallel()

	service := newTestEventService()
	userID := uuid.New()

	_, err := service.CreateEvent(context.Background(), &eventv1.CreateEventRequest{
		Event: newTestEvent(userID, "first", time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)

	_, err = service.CreateEvent(context.Background(), &eventv1.CreateEventRequest{
		Event: newTestEvent(userID, "second", time.Date(2026, 7, 1, 10, 30, 0, 0, time.UTC)),
	})

	require.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestListEventsStorageError(t *testing.T) {
	t.Parallel()

	service := newEventService(failingApp{})
	_, err := service.ListEventsForDay(context.Background(), &eventv1.ListEventsRequest{
		Date: timestamppb.Now(),
	})

	require.Equal(t, codes.Internal, status.Code(err))
}

func newTestEventService() *eventService {
	calendar := app.New(memorystorage.New())
	return newEventService(calendar)
}

func newTestEvent(userID uuid.UUID, title string, date time.Time) *eventv1.Event {
	return &eventv1.Event{
		Title:        title,
		Date:         timestamppb.New(date),
		Duration:     durationpb.New(time.Hour),
		Description:  wrapperspb.String("calendar api"),
		UserId:       userID.String(),
		NotifyBefore: durationpb.New(24 * time.Hour),
	}
}

type failingApp struct{}

func (failingApp) CreateEvent(context.Context, storage.Event) (storage.Event, error) {
	return storage.Event{}, errors.New("storage unavailable")
}

func (failingApp) UpdateEvent(context.Context, uuid.UUID, storage.Event) error {
	return errors.New("storage unavailable")
}

func (failingApp) DeleteEvent(context.Context, uuid.UUID) error {
	return errors.New("storage unavailable")
}

func (failingApp) ListEventsForDay(context.Context, time.Time) ([]storage.Event, error) {
	return nil, errors.New("storage unavailable")
}

func (failingApp) ListEventsForWeek(context.Context, time.Time) ([]storage.Event, error) {
	return nil, errors.New("storage unavailable")
}

func (failingApp) ListEventsForMonth(context.Context, time.Time) ([]storage.Event, error) {
	return nil, errors.New("storage unavailable")
}
