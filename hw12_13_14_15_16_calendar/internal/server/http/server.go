package internalhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
)

type Server struct {
	logger logger.Logger
	server *http.Server
}

type Application interface{}

func NewServer(logger logger.Logger, host string, port int, app Application) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)

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

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello World!"))
}
