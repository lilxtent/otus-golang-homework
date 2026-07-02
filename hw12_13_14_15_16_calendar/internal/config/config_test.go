package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewCalendar(t *testing.T) {
	t.Parallel()

	path := writeTempConfig(t, `
logger:
  level: INFO
http:
  host: 127.0.0.1
  port: 8080
grpc:
  host: 127.0.0.1
  port: 50051
storage:
  type: MEMORY
  dsn: postgres://postgres:password@localhost:5435/backend?sslmode=disable
`)

	config, err := NewCalendar(path)
	require.NoError(t, err)
	require.Equal(t, "INFO", config.Logger.Level)
	require.Equal(t, "127.0.0.1", config.HTTP.Host)
	require.Equal(t, 8080, config.HTTP.Port)
	require.Equal(t, "127.0.0.1", config.GRPC.Host)
	require.Equal(t, 50051, config.GRPC.Port)
	require.Equal(t, StorageMemory, config.Storage.Type)
	require.Equal(t, "postgres://postgres:password@localhost:5435/backend?sslmode=disable", config.Storage.DSN)
}

func TestNewScheduler(t *testing.T) {
	t.Parallel()

	path := writeTempConfig(t, `
logger:
  level: INFO
storage:
  type: SQL
  dsn: postgres://postgres:password@localhost:5435/backend?sslmode=disable
queue:
  url: amqp://rabbit:password@localhost:5672/
  exchange: calendar
  queue: calendar.notifications
  routing_key: calendar.notification
  consumer_tag: calendar-scheduler
scheduler:
  scan_interval: 30s
  cleanup_interval: 2h
`)

	config, err := NewScheduler(path)
	require.NoError(t, err)
	require.Equal(t, "INFO", config.Logger.Level)
	require.Equal(t, StorageSQL, config.Storage.Type)
	require.Equal(t, "postgres://postgres:password@localhost:5435/backend?sslmode=disable", config.Storage.DSN)
	require.Equal(t, "amqp://rabbit:password@localhost:5672/", config.Queue.URL)
	require.Equal(t, "calendar", config.Queue.Exchange)
	require.Equal(t, "calendar.notifications", config.Queue.Queue)
	require.Equal(t, "calendar.notification", config.Queue.RoutingKey)
	require.Equal(t, "calendar-scheduler", config.Queue.ConsumerTag)
	require.Equal(t, 30*time.Second, config.Scheduler.ScanInterval)
	require.Equal(t, 2*time.Hour, config.Scheduler.CleanupInterval)
}

func TestLoadReturnsErrorForMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.yaml")

	err := Load(path, &Calendar{})
	require.Error(t, err)
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}
