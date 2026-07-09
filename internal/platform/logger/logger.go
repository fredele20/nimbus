package logger

import (
	"log/slog"
	"os"

	"github.com/fredele20/nimbus/internal/platform/config"
)

func New(cfg config.LoggerConfig) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: cfg.Level == "debug",
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler)
}
