package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

func Load() (Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Ignore missing .env file.
	// This allows configuration to come entirely from
	// environment variables in production.
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError

		if !errors.As(err, &configFileNotFound) {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	setDefaults()

	cfg := Config{
		App: AppConfig{
			Env: viper.GetString(EnvApp),
		},
		Server: ServerConfig{
			Host: viper.GetString(EnvAppHost),
			Port: viper.GetString(EnvAppPort),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString(EnvDBHost),
			Port:     viper.GetInt(EnvDBPort),
			User:     viper.GetString(EnvDBUser),
			Password: viper.GetString(EnvDBPassword),
			Name:     viper.GetString(EnvDBName),
			SSLMode:  viper.GetString(EnvDBSSLMode),
		},
		Redis: RedisConfig{
			Host: viper.GetString(EnvRedisHost),
			Port: viper.GetInt(EnvRedisPort),
		},
		Logger: LoggerConfig{
			Level: viper.GetString(EnvLogLevel),
		},
		JWT: JWTConfig{
			Secret: viper.GetString(EnvJWTSecret),
		},
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func setDefaults() {
	viper.SetDefault(EnvAppHost, "0.0.0.0")
	viper.SetDefault(EnvAppPort, "8080")
	viper.SetDefault(EnvApp, "development")

	viper.SetDefault(EnvDBHost, "localhost")
	viper.SetDefault(EnvDBPort, 5432)
	viper.SetDefault(EnvDBUser, "postgres")
	viper.SetDefault(EnvDBPassword, "postgres")
	viper.SetDefault(EnvDBName, "nimbus")
	viper.SetDefault(EnvDBSSLMode, "disable")

	viper.SetDefault(EnvRedisHost, "localhost")
	viper.SetDefault(EnvRedisPort, 6379)

	viper.SetDefault(EnvLogLevel, "debug")
}

func validate(cfg Config) error {
	if cfg.Server.Port == "" {
		return errors.New("server port is required")
	}

	if cfg.Database.Host == "" {
		return errors.New("database host is required")
	}

	if cfg.Database.Name == "" {
		return errors.New("database name is required")
	}

	if cfg.Redis.Host == "" {
		return errors.New("redis host is required")
	}

	if cfg.JWT.Secret == "" {
		return errors.New("jwt secret is required")
	}

	switch cfg.Logger.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf(
			"invalid log level %q: must be one of debug, info, warn, error",
			cfg.Logger.Level,
		)
	}

	return nil
}
