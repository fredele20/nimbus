package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/fredele20/nimbus/internal/platform/config"
	"github.com/fredele20/nimbus/internal/platform/logger"
	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(r *gin.Engine, cfg config.Config) *Server {
	return &Server{
		logger: logger.New(cfg),
		httpServer: &http.Server{
			Addr:         ":" + cfg.Server.Port,
			Handler:      r,
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
