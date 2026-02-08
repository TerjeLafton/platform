package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/handlers"
	"github.com/terjelafton/platform/apps/web/internal/middleware"
)

func main() {
	cfg := LoadConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatal("failed to connect to NATS:", err)
	}
	defer nc.Drain()

	logger.Info("connected to NATS", "server", nc.ConnectedUrl())

	cookieName := "auth_token"
	requireAuth := middleware.RequireAuth(nc, cookieName, logger)

	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public routes
	mux.HandleFunc("GET /login", handlers.HandleLoginPage)
	mux.HandleFunc("POST /login", handlers.HandleLogin(nc, cookieName, logger))
	mux.HandleFunc("GET /register", handlers.HandleRegisterPage)
	mux.HandleFunc("POST /register", handlers.HandleRegister(nc, cookieName, logger))

	// Protected routes
	mux.Handle("GET /{$}", requireAuth(handlers.HandleListsPage(nc, logger)))
	mux.Handle("POST /logout", requireAuth(handlers.HandleLogout(cookieName)))
	mux.Handle("POST /lists", requireAuth(handlers.HandleCreateList(nc, logger)))
	mux.Handle("GET /lists/{id}", requireAuth(handlers.HandleListDetail(nc, logger)))
	mux.Handle("DELETE /lists/{id}", requireAuth(handlers.HandleDeleteList(nc, logger)))
	mux.Handle("POST /lists/{id}/title", requireAuth(handlers.HandleUpdateListTitle(nc, logger)))
	mux.Handle("POST /lists/{id}/items", requireAuth(handlers.HandleCreateItem(nc, logger)))
	mux.Handle("POST /items/{id}/toggle", requireAuth(handlers.HandleToggleItem(nc, logger)))
	mux.Handle("DELETE /items/{id}", requireAuth(handlers.HandleDeleteItem(nc, logger)))

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		logger.Info("web service starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed:", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down")
	srv.Shutdown(context.Background())
}
