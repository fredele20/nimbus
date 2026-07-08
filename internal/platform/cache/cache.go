package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fredele20/nimbus/internal/platform/config"
	"github.com/redis/go-redis/v9"
)

func New(cfg config.RedisConfig, log *slog.Logger) (*redis.Client, error) {
	opts := DefaultOptions()

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address(),
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolSize:     opts.PoolSize,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	log.Info("redis connected")

	return client, nil
}
