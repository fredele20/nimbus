package config

import "fmt"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	Host string
	Port int
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host,
		d.Port,
		d.User,
		d.Password,
		d.Name,
		d.SSLMode,
	)
}

type RedisConfig struct {
	Host string
	Port int
}

func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type LoggerConfig struct {
	Level string
}
