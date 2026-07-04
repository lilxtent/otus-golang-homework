package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/http/models"
	"github.com/google/uuid"
)

func main() {
	apiURL := flag.String("api", "http://localhost:8080", "Calendar HTTP API base URL")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now().UTC()
	description := "created by tools/create_events.go"

	events := []models.EventRequest{
		newEvent("reminder in 30s #1", now.Add(90*time.Second), "30m", &description, ptr("1m")),
		newEvent("reminder in 30s #2", now.Add(2*time.Minute), "45m", &description, ptr("90s")),
		newEvent("reminder later today", now.Add(2*time.Hour), "1h", &description, ptr("30m")),
		newEvent("no reminder", now.Add(3*time.Hour), "1h", &description, nil),
		newEvent("tomorrow planning", now.Add(24*time.Hour), "2h", &description, ptr("2h")),
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, event := range events {
		created, err := createEvent(ctx, client, *apiURL, event)
		if err != nil {
			fmt.Printf("create %q: %v\n", event.Title, err)
			continue
		}

		fmt.Printf(
			"created: id=%s title=%q date=%s user_id=%s notify_before=%s\n",
			created.ID,
			created.Title,
			created.Date,
			created.UserID,
			formatOptional(created.NotifyBefore),
		)
	}
}

func newEvent(
	title string,
	date time.Time,
	duration string,
	description *string,
	notifyBefore *string,
) models.EventRequest {
	return models.EventRequest{
		Title:        title,
		Date:         date.Format(time.RFC3339),
		Duration:     duration,
		Description:  description,
		UserID:       uuid.NewString(),
		NotifyBefore: notifyBefore,
	}
}

func createEvent(
	ctx context.Context,
	client *http.Client,
	apiURL string,
	event models.EventRequest,
) (models.EventResponse, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return models.EventResponse{}, err
	}

	endpoint := strings.TrimRight(apiURL, "/") + "/events"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return models.EventResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return models.EventResponse{}, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return models.EventResponse{}, err
	}
	if response.StatusCode != http.StatusCreated {
		return models.EventResponse{}, fmt.Errorf(
			"unexpected status %s: %s",
			response.Status,
			strings.TrimSpace(string(responseBody)),
		)
	}

	var created models.EventResponse
	if err := json.Unmarshal(responseBody, &created); err != nil {
		return models.EventResponse{}, err
	}

	return created, nil
}

func ptr(value string) *string {
	return &value
}

func formatOptional(value *string) string {
	if value == nil {
		return "<none>"
	}

	return *value
}
