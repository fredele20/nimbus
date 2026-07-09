package app

import (
	"log/slog"

	"github.com/fredele20/nimbus/internal/platform/config"
	"github.com/fredele20/nimbus/internal/platform/httpserver"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *gorm.DB
	Redis  *redis.Client
	Server *httpserver.Server
}
