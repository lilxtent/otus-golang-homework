package internalgrpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	eventv1 "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/grpc/pb/event/v1"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Server struct {
	logger logger.Logger
	server *grpc.Server
	addr   string
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) (storage.Event, error)
	UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storage.Event, error)
}

func NewServer(logger logger.Logger, host string, port int, app Application) *Server {
	server := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor(logger)))
	eventv1.RegisterEventServiceServer(server, newEventService(app))
	reflection.Register(server)

	return &Server{
		logger: logger,
		server: server,
		addr:   net.JoinHostPort(host, strconv.Itoa(port)),
	}
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		s.server.GracefulStop()
	}()

	s.logger.Info("grpc server is listening on " + s.addr)
	if err := s.server.Serve(listener); err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop(_ context.Context) error {
	s.logger.Info("stopping grpc server")
	s.server.GracefulStop()
	return nil
}

func loggingInterceptor(logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		code := status.Code(err)
		logger.Info(fmt.Sprintf(
			`%s [%s] %s %s %s`,
			clientAddr(ctx),
			start.Format("02/Jan/2006:15:04:05 -0700"),
			info.FullMethod,
			code.String(),
			time.Since(start),
		))

		return resp, err
	}
}

func clientAddr(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return "-"
	}
	return p.Addr.String()
}

type eventService struct {
	eventv1.UnimplementedEventServiceServer
	app Application
}

func newEventService(app Application) *eventService {
	return &eventService{app: app}
}

func (s *eventService) CreateEvent(
	ctx context.Context,
	req *eventv1.CreateEventRequest,
) (*eventv1.EventResponse, error) {
	event, err := toStorageEvent(req.GetEvent())
	if err != nil {
		return nil, err
	}

	created, err := s.app.CreateEvent(ctx, event)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &eventv1.EventResponse{Event: fromStorageEvent(created)}, nil
}

func (s *eventService) UpdateEvent(
	ctx context.Context,
	req *eventv1.UpdateEventRequest,
) (*eventv1.EventResponse, error) {
	id, err := parseUUID(req.GetId(), "id")
	if err != nil {
		return nil, err
	}

	event, err := toStorageEvent(req.GetEvent())
	if err != nil {
		return nil, err
	}

	if err := s.app.UpdateEvent(ctx, id, event); err != nil {
		return nil, toStatusError(err)
	}

	event.ID = id
	return &eventv1.EventResponse{Event: fromStorageEvent(event)}, nil
}

func (s *eventService) DeleteEvent(ctx context.Context, req *eventv1.DeleteEventRequest) (*emptypb.Empty, error) {
	id, err := parseUUID(req.GetId(), "id")
	if err != nil {
		return nil, err
	}

	if err := s.app.DeleteEvent(ctx, id); err != nil {
		return nil, toStatusError(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *eventService) ListEventsForDay(
	ctx context.Context,
	req *eventv1.ListEventsRequest,
) (*eventv1.ListEventsResponse, error) {
	return s.listEvents(ctx, req, s.app.ListEventsForDay)
}

func (s *eventService) ListEventsForWeek(
	ctx context.Context,
	req *eventv1.ListEventsRequest,
) (*eventv1.ListEventsResponse, error) {
	return s.listEvents(ctx, req, s.app.ListEventsForWeek)
}

func (s *eventService) ListEventsForMonth(
	ctx context.Context,
	req *eventv1.ListEventsRequest,
) (*eventv1.ListEventsResponse, error) {
	return s.listEvents(ctx, req, s.app.ListEventsForMonth)
}

func (s *eventService) listEvents(
	ctx context.Context,
	req *eventv1.ListEventsRequest,
	list func(context.Context, time.Time) ([]storage.Event, error),
) (*eventv1.ListEventsResponse, error) {
	if req.GetDate() == nil || !req.GetDate().IsValid() {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}

	events, err := list(ctx, req.GetDate().AsTime())
	if err != nil {
		return nil, toStatusError(err)
	}

	response := &eventv1.ListEventsResponse{Events: make([]*eventv1.Event, 0, len(events))}
	for _, event := range events {
		response.Events = append(response.Events, fromStorageEvent(event))
	}
	return response, nil
}

func toStorageEvent(event *eventv1.Event) (storage.Event, error) {
	if event == nil {
		return storage.Event{}, status.Error(codes.InvalidArgument, "event is required")
	}
	if event.GetTitle() == "" {
		return storage.Event{}, status.Error(codes.InvalidArgument, "title is required")
	}
	if event.GetDate() == nil || !event.GetDate().IsValid() {
		return storage.Event{}, status.Error(codes.InvalidArgument, "date is required")
	}
	if event.GetDuration() == nil || !event.GetDuration().IsValid() {
		return storage.Event{}, status.Error(codes.InvalidArgument, "duration is required")
	}

	userID, err := parseUUID(event.GetUserId(), "user_id")
	if err != nil {
		return storage.Event{}, err
	}

	result := storage.Event{
		Title:    event.GetTitle(),
		Date:     event.GetDate().AsTime(),
		Duration: event.GetDuration().AsDuration(),
		UserID:   userID,
	}

	if event.GetId() != "" {
		id, err := parseUUID(event.GetId(), "event.id")
		if err != nil {
			return storage.Event{}, err
		}
		result.ID = id
	}
	if event.GetDescription() != nil {
		description := event.GetDescription().GetValue()
		result.Description = &description
	}
	if event.GetNotifyBefore() != nil {
		if !event.GetNotifyBefore().IsValid() {
			return storage.Event{}, status.Error(codes.InvalidArgument, "notify_before is invalid")
		}
		notifyBefore := event.GetNotifyBefore().AsDuration()
		result.NotifyBefore = &notifyBefore
	}

	return result, nil
}

func fromStorageEvent(event storage.Event) *eventv1.Event {
	result := &eventv1.Event{
		Id:       event.ID.String(),
		Title:    event.Title,
		Date:     timestamppb.New(event.Date),
		Duration: durationpb.New(event.Duration),
		UserId:   event.UserID.String(),
	}
	if event.Description != nil {
		result.Description = wrapperspb.String(*event.Description)
	}
	if event.NotifyBefore != nil {
		result.NotifyBefore = durationpb.New(*event.NotifyBefore)
	}
	return result
}

func parseUUID(value, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, status.Errorf(codes.InvalidArgument, "%s must be a valid uuid", field)
	}
	return id, nil
}

func toStatusError(err error) error {
	switch {
	case errors.Is(err, storage.ErrEventAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, storage.ErrDateBusy):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, storage.ErrEventNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
