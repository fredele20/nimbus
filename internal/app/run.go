package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fredele20/nimbus/internal/platform/cache"
	"github.com/fredele20/nimbus/internal/platform/config"
	"github.com/fredele20/nimbus/internal/platform/database"
	"github.com/fredele20/nimbus/internal/platform/httpserver"
	"github.com/fredele20/nimbus/internal/platform/logger"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Logger)

	db, err := database.New(cfg.Database, log)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	redisClient, err := cache.New(cfg.Redis, log)

	server := httpserver.New(cfg.Server, log)

	app := &App{
		Config: &cfg,
		Logger: log,
		DB:     db,
		Redis:  redisClient,
		Server: server,
	}

	go func() {
		if err := app.Server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		return err
	}

	if err := sqlDB.Close(); err != nil {
		app.Logger.Warn("failed to close database", "error", err)
	}

	if err := app.Redis.Close(); err != nil {
		app.Logger.Warn("failed to close redis", "error", err)
	}

	app.Logger.Info("application stopped")

	return nil
}
