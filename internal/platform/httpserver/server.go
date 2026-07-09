package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/fredele20/nimbus/internal/platform/config"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(cfg config.ServerConfig, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(
			`{"status":"ok","service":"nimbus","version":"dev","timestamp":"%s"}`,
			time.Now().UTC().Format(time.RFC3339),
		)))
	})

	return &Server{
		logger: logger,
		httpServer: &http.Server{
			Addr:         cfg.Address(),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	s.logger.Info("starting HTTP server", "address", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}
