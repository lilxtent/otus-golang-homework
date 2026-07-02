package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/http/models"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateAndListEvents(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	userID := uuid.New()
	body := `{
		"title": "demo",
		"date": "2026-07-01T10:00:00Z",
		"duration": "1h",
		"description": "calendar api",
		"userId": "` + userID.String() + `",
		"notifyBefore": "24h"
	}`

	response := performRequest(handler, http.MethodPost, "/events", body)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())

	var created models.EventResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&created))
	require.NotEqual(t, uuid.Nil.String(), created.ID)

	response = performRequest(handler, http.MethodGet, "/events/day?date=2026-07-01", "")
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	var events []models.EventResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&events))
	require.Len(t, events, 1)
	require.Equal(t, created.ID, events[0].ID)
	require.Equal(t, "demo", events[0].Title)
	require.Equal(t, "1h0m0s", events[0].Duration)
}

func TestUpdateAndDeleteEvent(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	userID := uuid.New()
	createBody := `{
		"title": "demo",
		"date": "2026-07-01T10:00:00Z",
		"duration": "1h",
		"userId": "` + userID.String() + `"
	}`
	updateBody := `{
		"title": "updated",
		"date": "2026-07-02T10:00:00Z",
		"duration": "2h",
		"userId": "` + userID.String() + `"
	}`

	response := performRequest(handler, http.MethodPost, "/events", createBody)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())

	var created models.EventResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&created))
	eventID := uuid.MustParse(created.ID)

	response = performRequest(handler, http.MethodPut, "/events/"+eventID.String(), updateBody)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	response = performRequest(handler, http.MethodDelete, "/events/"+eventID.String(), "")
	require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())

	response = performRequest(handler, http.MethodDelete, "/events/"+eventID.String(), "")
	require.Equal(t, http.StatusNotFound, response.Code, response.Body.String())
}

func TestUpdateEventValidation(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	userID := uuid.New()
	body := `{
		"title": "updated",
		"date": "2026-07-02T10:00:00Z",
		"duration": "2h",
		"userId": "` + userID.String() + `"
	}`

	response := performRequest(handler, http.MethodPut, "/events/not-a-uuid", body)

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
}

func TestDeleteEventValidation(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	response := performRequest(handler, http.MethodDelete, "/events/not-a-uuid", "")

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
}

func TestCreateEventValidation(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	response := performRequest(handler, http.MethodPost, "/events", `{"title": "bad"}`)

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
}

func TestCreateEventDateBusy(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	userID := uuid.New()
	first := `{
		"title": "first",
		"date": "2026-07-01T10:00:00Z",
		"duration": "1h",
		"userId": "` + userID.String() + `"
	}`
	second := `{
		"title": "second",
		"date": "2026-07-01T10:30:00Z",
		"duration": "1h",
		"userId": "` + userID.String() + `"
	}`

	response := performRequest(handler, http.MethodPost, "/events", first)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())

	response = performRequest(handler, http.MethodPost, "/events", second)
	require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
}

func TestListEventsStorageError(t *testing.T) {
	t.Parallel()

	handler := newMux(failingApp{})
	response := performRequest(handler, http.MethodGet, "/events/day?date=2026-07-01", "")

	require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
}

func newTestHandler() http.Handler {
	calendar := app.New(memorystorage.New())
	return newMux(calendar)
}

func performRequest(handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
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
