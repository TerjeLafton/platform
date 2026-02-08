package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"

	"github.com/terjelafton/platform/apps/todo/internal/db"
	natshandler "github.com/terjelafton/platform/apps/todo/internal/handlers/nats"
	"github.com/terjelafton/platform/apps/todo/internal/service"
)

func main() {
	cfg := LoadConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	dbConn, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatal("failed to ping database:", err)
	}

	queries := db.New(dbConn)
	logger.Info("connected to database")

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatal("failed to connect to NATS:", err)
	}
	defer func() {
		logger.Info("draining NATS connection...")
		nc.Drain()
		logger.Info("todo service stopped")
	}()
	logger.Info("connected to NATS")

	svc := service.New(queries, nc, logger)

	handler := natshandler.New(svc, logger)
	if err := handler.Register(nc); err != nil {
		log.Fatal("failed to register NATS handlers:", err)
	}

	logger.Info("todo service ready",
		"nats_url", nc.ConnectedUrl(),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("shutdown signal received, cleaning up...")
}
