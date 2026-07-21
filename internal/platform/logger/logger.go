package logger

import (
	"log/slog"
	"os"

	"github.com/fredele20/nimbus/internal/platform/config"
	"github.com/lmittmann/tint"
)

func New(cfg config.Config) *slog.Logger {
	var logger *slog.Logger
	opts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Logger.Level),
		AddSource: cfg.Logger.Level == "debug",
	}

	if cfg.App.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level:      parseLevel(cfg.Logger.Level),
			AddSource:  true,
			TimeFormat: "15:04:05",
		}))
	}

	return logger
}
