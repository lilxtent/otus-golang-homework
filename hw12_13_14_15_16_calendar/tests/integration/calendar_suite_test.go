//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/http/models"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultAPIURL    = "http://calendar:8080"
	defaultRabbitURL = "amqp://rabbit:password@rabbitmq:5672/"

	statusExchange   = "calendar"
	statusQueue      = "calendar.notification_statuses"
	statusRoutingKey = "calendar.notification.status"
)

var (
	apiURL string
	client *http.Client

	rabbitConn *amqp.Connection
	rabbitCh   *amqp.Channel
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Calendar Integration Suite")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	apiURL = envOrDefault("CALENDAR_API_URL", defaultAPIURL)
	client = &http.Client{Timeout: 5 * time.Second}

	Eventually(func(g Gomega) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/events/day?date=2026-01-01", nil)
		g.Expect(err).NotTo(HaveOccurred())

		resp, err := client.Do(req)
		g.Expect(err).NotTo(HaveOccurred())
		defer closeBody(resp.Body)

		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	}, 60*time.Second, time.Second).Should(Succeed())

	var err error
	rabbitConn, err = amqp.Dial(envOrDefault("RABBITMQ_URL", defaultRabbitURL))
	Expect(err).NotTo(HaveOccurred())

	rabbitCh, err = rabbitConn.Channel()
	Expect(err).NotTo(HaveOccurred())
	Expect(declareStatusQueue(rabbitCh)).To(Succeed())
})

var _ = AfterSuite(func() {
	if rabbitCh != nil {
		_ = rabbitCh.Close()
	}
	if rabbitConn != nil {
		_ = rabbitConn.Close()
	}
})

var _ = Describe("Calendar HTTP API", func() {
	It("creates an event and returns a business error for overlapping events", func(ctx SpecContext) {
		userID := uuid.NewString()
		start := futureDate(48 * time.Hour)

		created := createEvent(ctx, models.EventRequest{
			Title:    "integration overlap source",
			Date:     start.Format(time.RFC3339),
			Duration: "1h",
			UserID:   userID,
		}, http.StatusCreated)

		Expect(created.ID).NotTo(BeEmpty())
		Expect(created.UserID).To(Equal(userID))

		conflictBody := createEventRaw(ctx, models.EventRequest{
			Title:    "integration overlap conflict",
			Date:     start.Add(30 * time.Minute).Format(time.RFC3339),
			Duration: "30m",
			UserID:   userID,
		}, http.StatusConflict)
		Expect(conflictBody).To(ContainSubstring("date is busy"))
	})

	It("lists events for day, week, and month", func(ctx SpecContext) {
		userID := uuid.NewString()
		day := futureDate(7 * 24 * time.Hour)
		week := day.AddDate(0, 0, 2)
		month := day.AddDate(0, 0, 10)

		created := []models.EventResponse{
			createEvent(ctx, models.EventRequest{Title: "integration day", Date: day.Format(time.RFC3339), Duration: "30m", UserID: userID}, http.StatusCreated),
			createEvent(ctx, models.EventRequest{Title: "integration week", Date: week.Format(time.RFC3339), Duration: "30m", UserID: userID}, http.StatusCreated),
			createEvent(ctx, models.EventRequest{Title: "integration month", Date: month.Format(time.RFC3339), Duration: "30m", UserID: userID}, http.StatusCreated),
		}

		dayEvents := listEvents(ctx, "/events/day", day)
		Expect(eventIDs(dayEvents)).To(ContainElement(created[0].ID))

		weekEvents := listEvents(ctx, "/events/week", day)
		Expect(eventIDs(weekEvents)).To(ContainElements(created[0].ID, created[1].ID))

		monthEvents := listEvents(ctx, "/events/month", day)
		Expect(eventIDs(monthEvents)).To(ContainElements(created[0].ID, created[1].ID, created[2].ID))
	})
})

var _ = Describe("Calendar notifications", func() {
	It("publishes sender delivery status after a due notification is processed", func(ctx SpecContext) {
		Expect(declareStatusQueue(rabbitCh)).To(Succeed())
		_, err := rabbitCh.QueuePurge(statusQueue, false)
		Expect(err).NotTo(HaveOccurred())

		deliveries, err := rabbitCh.ConsumeWithContext(ctx, statusQueue, "integration-status", true, false, false, false, nil)
		Expect(err).NotTo(HaveOccurred())

		notifyBefore := "1m"
		created := createEvent(ctx, models.EventRequest{
			Title:        "integration notification",
			Date:         time.Now().UTC().Add(10 * time.Second).Format(time.RFC3339),
			Duration:     "30m",
			UserID:       uuid.NewString(),
			NotifyBefore: &notifyBefore,
		}, http.StatusCreated)

		Eventually(func(g Gomega) storage.NotificationStatus {
			select {
			case delivery := <-deliveries:
				var status storage.NotificationStatus
				g.Expect(json.Unmarshal(delivery.Body, &status)).To(Succeed())
				if status.EventID != created.ID {
					return storage.NotificationStatus{}
				}
				return status
			default:
				return storage.NotificationStatus{}
			}
		}, 30*time.Second, 500*time.Millisecond).Should(SatisfyAll(
			WithTransform(func(status storage.NotificationStatus) string { return status.EventID }, Equal(created.ID)),
			WithTransform(func(status storage.NotificationStatus) string { return status.Status }, Equal("sent")),
			WithTransform(func(status storage.NotificationStatus) bool { return status.SentAt.IsZero() }, BeFalse()),
		))
	})
})

func createEvent(ctx context.Context, request models.EventRequest, expectedStatus int) models.EventResponse {
	GinkgoHelper()

	body := createEventRaw(ctx, request, expectedStatus)
	if expectedStatus != http.StatusCreated {
		return models.EventResponse{}
	}

	var response models.EventResponse
	Expect(json.Unmarshal([]byte(body), &response)).To(Succeed())
	return response
}

func createEventRaw(ctx context.Context, request models.EventRequest, expectedStatus int) string {
	GinkgoHelper()

	payload, err := json.Marshal(request)
	Expect(err).NotTo(HaveOccurred())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/events", bytes.NewReader(payload))
	Expect(err).NotTo(HaveOccurred())
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	Expect(err).NotTo(HaveOccurred())
	defer closeBody(resp.Body)

	body, err := io.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(expectedStatus), strings.TrimSpace(string(body)))

	return string(body)
}

func listEvents(ctx context.Context, path string, date time.Time) []models.EventResponse {
	GinkgoHelper()

	url := fmt.Sprintf("%s%s?date=%s", apiURL, path, date.Format(time.DateOnly))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	Expect(err).NotTo(HaveOccurred())

	resp, err := client.Do(req)
	Expect(err).NotTo(HaveOccurred())
	defer closeBody(resp.Body)

	body, err := io.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK), strings.TrimSpace(string(body)))

	var events []models.EventResponse
	Expect(json.Unmarshal(body, &events)).To(Succeed())
	return events
}

func eventIDs(events []models.EventResponse) []string {
	ids := make([]string, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}

	return ids
}

func declareStatusQueue(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(statusExchange, amqp.ExchangeDirect, true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(statusQueue, true, false, false, false, nil); err != nil {
		return err
	}

	return ch.QueueBind(statusQueue, statusRoutingKey, statusExchange, false, nil)
}

func futureDate(offset time.Duration) time.Time {
	now := time.Now().UTC().Add(offset)
	return time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func closeBody(body io.Closer) {
	_ = body.Close()
}
