package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/http/models"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	logger logger.Logger
	server *http.Server
}

var validate = validator.New()

func NewServer(logger logger.Logger, host string, port int, application app.Application) *Server {
	mux := newMux(application)

	return &Server{
		logger: logger,
		server: &http.Server{
			Addr:              net.JoinHostPort(host, strconv.Itoa(port)),
			Handler:           loggingMiddleware(logger, mux),
			ReadHeaderTimeout: time.Second * 5,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	_ = ctx

	s.logger.Info("http server is listening on " + s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping http server")
	return s.server.Shutdown(ctx)
}

func newMux(app app.Application) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /events", createEventHandler(app))
	mux.HandleFunc("PUT /events/{id}", updateEventHandler(app))
	mux.HandleFunc("DELETE /events/{id}", deleteEventHandler(app))
	mux.HandleFunc("GET /events/day", listEventsHandler(app.ListEventsForDay))
	mux.HandleFunc("GET /events/week", listEventsHandler(app.ListEventsForWeek))
	mux.HandleFunc("GET /events/month", listEventsHandler(app.ListEventsForMonth))
	mux.HandleFunc("/", notFoundHandler)
	return mux
}

func createEventHandler(app app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeCreateEventRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		event, err := request.ToEvent()
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		event, err = app.CreateEvent(r.Context(), event)
		if err != nil {
			writeStorageError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, models.NewEventResponse(event))
	}
}

func updateEventHandler(app app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeUpdateEventRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		id := request.EventID()
		event, err := request.ToEvent()
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		if err := app.UpdateEvent(r.Context(), id, event); err != nil {
			writeStorageError(w, err)
			return
		}

		event.ID = id
		writeJSON(w, http.StatusOK, models.NewEventResponse(event))
	}
}

func deleteEventHandler(app app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeDeleteEventRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		if err := app.DeleteEvent(r.Context(), request.EventID()); err != nil {
			writeStorageError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func listEventsHandler(list func(context.Context, time.Time) ([]storage.Event, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date, err := parseDateParam(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		events, err := list(r.Context(), date)
		if err != nil {
			writeStorageError(w, err)
			return
		}

		response := make([]models.EventResponse, 0, len(events))
		for _, event := range events {
			response = append(response, models.NewEventResponse(event))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, errors.New("not found"))
}

func decodeCreateEventRequest(r *http.Request) (models.EventRequest, error) {
	defer func() {
		_ = r.Body.Close()
	}()

	var request models.EventRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return models.EventRequest{}, err
	}

	if err := validate.Struct(request); err != nil {
		return models.EventRequest{}, err
	}

	return request, nil
}

func decodeUpdateEventRequest(r *http.Request) (models.UpdateEventRequest, error) {
	eventRequest, err := decodeCreateEventRequest(r)
	if err != nil {
		return models.UpdateEventRequest{}, err
	}

	request := models.UpdateEventRequest{
		ID:           r.PathValue("id"),
		EventRequest: eventRequest,
	}
	if err := validate.Struct(request); err != nil {
		return models.UpdateEventRequest{}, err
	}

	return request, nil
}

func decodeDeleteEventRequest(r *http.Request) (models.DeleteEventRequest, error) {
	request := models.DeleteEventRequest{ID: r.PathValue("id")}
	if err := validate.Struct(request); err != nil {
		return models.DeleteEventRequest{}, err
	}

	return request, nil
}

func parseDateParam(r *http.Request) (time.Time, error) {
	value := r.URL.Query().Get("date")
	if value == "" {
		return time.Time{}, errors.New("date query parameter is required")
	}

	date, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return date, nil
	}

	date, err = time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, errors.New("date must be RFC3339 or YYYY-MM-DD")
	}
	return date, nil
}

func writeStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrEventAlreadyExists):
		writeError(w, http.StatusConflict, err)
	case errors.Is(err, storage.ErrDateBusy):
		writeError(w, http.StatusConflict, err)
	case errors.Is(err, storage.ErrEventNotFound):
		writeError(w, http.StatusNotFound, err)
	default:
		writeError(w, http.StatusInternalServerError, errors.New("internal server error"))
	}
}

func writeError(w http.ResponseWriter, statusCode int, err error) {
	writeJSON(w, statusCode, models.ErrorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
