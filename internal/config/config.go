package config

import (
	"errors"
	"os"
)

type Config struct {
	DBDSN     string
	HTTPPort  string
	JWTSecret string
}

// Load загружает конфиг из переменных окружения.
// DB_DSN — строка подключения к БД (по умолчанию: postgres://app:app@localhost:5432/cyber_risk?sslmode=disable)
// HTTP_PORT — порт HTTP сервера (по умолчанию: 8081)
// JWT_SECRET — секрет для подписи JWT. Обязателен в проде; dev-дефолт используется только при APP_ENV=dev (или пустом).
func Load() (*Config, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://app:app@localhost:5432/cyber_risk?sslmode=disable"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8081"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		env := os.Getenv("APP_ENV")
		if env != "" && env != "dev" && env != "development" {
			return nil, errors.New("JWT_SECRET is required when APP_ENV is not dev")
		}
		jwtSecret = "dev-secret-change-me"
	}

	return &Config{
		DBDSN:     dsn,
		HTTPPort:  port,
		JWTSecret: jwtSecret,
	}, nil
}
