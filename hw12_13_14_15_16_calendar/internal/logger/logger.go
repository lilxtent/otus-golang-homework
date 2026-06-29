package logger

import (
	"errors"
	"log/slog"
	"os"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}
type SLogger struct {
	logger *slog.Logger
}

func New(level string) (*SLogger, error) {
	logLevel, err := mapLogLevel(level)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	return &SLogger{
		logger: logger,
	}, nil
}

func (l *SLogger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *SLogger) Error(msg string) {
	l.logger.Error(msg)
}

func mapLogLevel(level string) (slog.Level, error) {
	switch level {
	case "debug", "DEBUG":
		return slog.LevelDebug, nil
	case "info", "INFO":
		return slog.LevelInfo, nil
	case "warn", "WARN":
		return slog.LevelWarn, nil
	case "error", "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelError, errors.New("unexpected log level " + level)
	}
}
