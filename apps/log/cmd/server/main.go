package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"

	"github.com/terjelafton/platform/apps/log/internal/db"
	natshandler "github.com/terjelafton/platform/apps/log/internal/handlers/nats"
	"github.com/terjelafton/platform/apps/log/internal/service"
)

func main() {
	cfg := LoadConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With("service", "log")

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
		logger.Info("log service stopped")
	}()
	logger.Info("connected to NATS")

	svc := service.New(queries, logger)

	handler := natshandler.New(svc, logger)
	if err := handler.Register(nc); err != nil {
		log.Fatal("failed to register NATS handlers:", err)
	}

	// Start retention cleanup goroutine (runs daily, deletes logs older than 30 days)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().AddDate(0, 0, -30)
			if err := svc.DeleteOldLogs(context.Background(), cutoff); err != nil {
				logger.Error("retention cleanup failed", "error", err)
			}
		}
	}()

	logger.Info("log service ready", "nats_url", nc.ConnectedUrl())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("shutdown signal received, cleaning up...")
}
