package database

import (
	"time"

	"gorm.io/gorm/logger"
)


type Options struct {
	MaxIdleConns int
	MaxOpenConns int
	ConnMaxLifetime time.Duration
	LogLevel logger.LogLevel
}

func DefaultOptions() Options {
	return Options{
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		ConnMaxLifetime: time.Hour,
		LogLevel: logger.Warn,
	}
}
