package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Parallel()

	path := writeTempConfig(t, `
logger:
  level: INFO
http:
  host: 127.0.0.1
  port: 8080
storage:
  type: MEMORY
  dsn: postgres://postgres:password@localhost:5435/backend?sslmode=disable
`)

	config, err := NewConfig(path)
	require.NoError(t, err)
	require.Equal(t, "INFO", config.Logger.Level)
	require.Equal(t, "127.0.0.1", config.HTTP.Host)
	require.Equal(t, 8080, config.HTTP.Port)
	require.Equal(t, StorageMemory, config.Storage.Type)
	require.Equal(t, "postgres://postgres:password@localhost:5435/backend?sslmode=disable", config.Storage.DSN)
}

func TestNewConfigReturnsErrorForMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := NewConfig(path)
	require.Error(t, err)
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}
