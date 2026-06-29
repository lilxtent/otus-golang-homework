package logger

import (
	"errors"
	"log/slog"
	"os"
	"strings"
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
	levelInUpperCase := strings.ToUpper(level)

	switch levelInUpperCase {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelError, errors.New("unexpected log level " + level)
	}
}
