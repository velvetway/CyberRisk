package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Diplom/internal/config"
	"Diplom/internal/repository"
	httpTransport "Diplom/internal/transport/http"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	// 2. Подключаемся к PostgreSQL через pgxpool
	db, err := repository.NewPostgresPool(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}
	defer db.Close()

	log.Println("connected to database")

	// 3. Создаём Fiber-приложение
	app := httpTransport.NewServer(ctx, db, cfg.JWTSecret)

	// 4. Запускаем сервер с graceful shutdown
	go func() {
		addr := ":" + cfg.HTTPPort
		log.Printf("starting HTTP server on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("http server stopped: %v", err)
			cancel()
		}
	}()

	// Ждём сигнала остановки
	<-ctx.Done()
	log.Println("shutdown signal received")

	// Даём немного времени на корректное завершение запросов
	_, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(); err != nil {
		log.Printf("error during http shutdown: %v", err)
	}

	log.Println("server gracefully stopped")
}
