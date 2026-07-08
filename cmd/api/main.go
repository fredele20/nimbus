package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fredele20/nimbus/internal/platform/config"
	"github.com/fredele20/nimbus/internal/platform/httpserver"
	"github.com/fredele20/nimbus/internal/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.Logger)

	server := httpserver.New(cfg.Server, log)

	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown server", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped")
}
