package config

import (
	"os"
)

type Config struct {
	DBDSN    string
	HTTPPort string
}

// Load загружает конфиг из переменных окружения.
// DB_DSN — строка подключения к БД (по умолчанию: postgres://app:app@localhost:5432/cyber_risk?sslmode=disable)
// HTTP_PORT — порт HTTP сервера (по умолчанию: 8081)
func Load() (*Config, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://app:app@localhost:5432/cyber_risk?sslmode=disable"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8081"
	}

	return &Config{
		DBDSN:    dsn,
		HTTPPort: port,
	}, nil
}
