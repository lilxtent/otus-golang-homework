package models

import (
	"errors"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type EventRequest struct {
	Title        string  `json:"title" validate:"required"`
	Date         string  `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Duration     string  `json:"duration" validate:"required"`
	Description  *string `json:"description"`
	UserID       string  `json:"userId" validate:"required,uuid"`
	NotifyBefore *string `json:"notifyBefore"`
}

type UpdateEventRequest struct {
	ID string `validate:"required,uuid"`
	EventRequest
}

type DeleteEventRequest struct {
	ID string `validate:"required,uuid"`
}

type EventResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Date         string  `json:"date"`
	Duration     string  `json:"duration"`
	Description  *string `json:"description,omitempty"`
	UserID       string  `json:"userId"`
	NotifyBefore *string `json:"notifyBefore,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (r EventRequest) ToEvent() (storage.Event, error) {
	var event storage.Event

	event.Title = r.Title

	date, _ := time.Parse(time.RFC3339, r.Date)
	event.Date = date

	duration, err := time.ParseDuration(r.Duration)
	if err != nil {
		return storage.Event{}, errors.New("duration must be a valid duration")
	}
	event.Duration = duration

	userID, _ := uuid.Parse(r.UserID)
	event.UserID = userID

	event.Description = r.Description
	if r.NotifyBefore != nil {
		notifyBefore, err := time.ParseDuration(*r.NotifyBefore)
		if err != nil {
			return storage.Event{}, errors.New("notifyBefore must be a valid duration")
		}
		event.NotifyBefore = &notifyBefore
	}

	return event, nil
}

func (r UpdateEventRequest) EventID() uuid.UUID {
	id, _ := uuid.Parse(r.ID)
	return id
}

func (r DeleteEventRequest) EventID() uuid.UUID {
	id, _ := uuid.Parse(r.ID)
	return id
}

func NewEventResponse(event storage.Event) EventResponse {
	response := EventResponse{
		ID:          event.ID.String(),
		Title:       event.Title,
		Date:        event.Date.Format(time.RFC3339),
		Duration:    event.Duration.String(),
		Description: event.Description,
		UserID:      event.UserID.String(),
	}
	if event.NotifyBefore != nil {
		notifyBefore := event.NotifyBefore.String()
		response.NotifyBefore = &notifyBefore
	}

	return response
}
