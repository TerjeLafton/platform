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
	})).With("service", "web")

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatal("failed to connect to NATS:", err)
	}
	defer nc.Drain()

	logger.Info("connected to NATS", "server", nc.ConnectedUrl())

	handlerLogger := logger.With("module", "handler")
	mwLogger := logger.With("module", "middleware")

	cookieName := "auth_token"
	requireAuth := middleware.RequireAuth(nc, cookieName, mwLogger)

	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public routes
	mux.HandleFunc("GET /login", handlers.HandleLoginPage)
	mux.HandleFunc("POST /login", handlers.HandleLogin(nc, cookieName, handlerLogger))
	mux.HandleFunc("GET /register", handlers.HandleRegisterPage)
	mux.HandleFunc("POST /register", handlers.HandleRegister(nc, cookieName, handlerLogger))
	mux.HandleFunc("GET /avatar/{id}", handlers.HandleAvatar(nc, handlerLogger))

	// Protected routes
	mux.Handle("GET /{$}", requireAuth(handlers.HandleHomePage(nc, handlerLogger)))
	mux.Handle("POST /logout", requireAuth(handlers.HandleLogout(cookieName)))
	mux.Handle("GET /profile", requireAuth(handlers.HandleProfilePage(nc, handlerLogger)))
	mux.Handle("POST /profile/avatar", requireAuth(handlers.HandleUploadAvatar(nc, handlerLogger)))
	mux.Handle("GET /todo/{$}", requireAuth(handlers.HandleListsPage(nc, handlerLogger)))
	mux.Handle("POST /todo", requireAuth(handlers.HandleCreateList(nc, handlerLogger)))
	mux.Handle("GET /todo/{id}", requireAuth(handlers.HandleListDetail(nc, handlerLogger)))
	mux.Handle("DELETE /todo/{id}", requireAuth(handlers.HandleDeleteList(nc, handlerLogger)))
	mux.Handle("POST /todo/{id}/title", requireAuth(handlers.HandleUpdateListTitle(nc, handlerLogger)))
	mux.Handle("POST /todo/{id}/items", requireAuth(handlers.HandleCreateItem(nc, handlerLogger)))
	mux.Handle("POST /todo/items/{id}/toggle", requireAuth(handlers.HandleToggleItem(nc, handlerLogger)))
	mux.Handle("DELETE /todo/items/{id}", requireAuth(handlers.HandleDeleteItem(nc, handlerLogger)))

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
