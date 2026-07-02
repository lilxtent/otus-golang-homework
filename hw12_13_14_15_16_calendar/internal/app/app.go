package app

import (
	"context"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type App struct {
	storage storage.Storage
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) (storage.Event, error)
	UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storage.Event, error)
}

func New(storage storage.Storage) *App {
	return &App{storage: storage}
}

func (a *App) CreateEvent(_ context.Context, event storage.Event) (storage.Event, error) {
	return a.storage.CreateEvent(event)
}

func (a *App) UpdateEvent(_ context.Context, id uuid.UUID, event storage.Event) error {
	return a.storage.UpdateEvent(id, event)
}

func (a *App) DeleteEvent(_ context.Context, id uuid.UUID) error {
	return a.storage.DeleteEvent(id)
}

func (a *App) ListEventsForDay(_ context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForDay(date)
}

func (a *App) ListEventsForWeek(_ context.Context, startOfWeek time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForWeek(startOfWeek)
}

func (a *App) ListEventsForMonth(_ context.Context, startOfMonth time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForMonth(startOfMonth)
}
