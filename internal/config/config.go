package config

import (
	"errors"
	"os"
)

type Config struct {
	DBDSN     string
	HTTPPort  string
	JWTSecret string

	// BDUSQLitePath — снимок БДУ ФСТЭК. Файл один, но читают его двое и
	// по-разному: bdu.Snapshot ищет уязвимости конкретной версии ПО для
	// автодетекта, а BDUSQLiteRepository отдаёт каталог с фильтрами по
	// CVSS и вендору. Это не дублирование, а разные задачи над общими
	// данными, поэтому путь тоже общий.
	BDUSQLitePath string

	// MinreestrSQLitePath — выборка каталога отечественного ПО.
	MinreestrSQLitePath string
	// SZISQLitePath — реестр сертифицированных средств защиты ФСТЭК.
	SZISQLitePath string
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

	// BDU_SNAPSHOT_PATH — прежнее имя переменной; поддерживается, чтобы
	// не ломать развёртывания, где оно уже прописано.
	bduPath := envOrDefault("BDU_SQLITE_PATH", "")
	if bduPath == "" {
		bduPath = envOrDefault("BDU_SNAPSHOT_PATH", "./data/bdu.sqlite")
	}

	return &Config{
		DBDSN:               dsn,
		HTTPPort:            port,
		JWTSecret:           jwtSecret,
		BDUSQLitePath:       bduPath,
		MinreestrSQLitePath: envOrDefault("MINREESTR_SQLITE_PATH", "./data/minreestr.sqlite"),
		SZISQLitePath:       envOrDefault("SZI_SQLITE_PATH", "./data/szi.sqlite"),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
