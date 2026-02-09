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
	"github.com/terjelafton/platform/apps/web/internal/sse"
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
	sseManager := sse.NewManager()

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
	mux.Handle("GET /todo/new", requireAuth(handlers.HandleCreateListPage(nc, handlerLogger)))
	mux.Handle("POST /todo/new", requireAuth(handlers.HandleCreateListFromForm(nc, handlerLogger)))
	mux.Handle("GET /todo/templates/{$}", requireAuth(handlers.HandleTemplatesPage(nc, handlerLogger)))
	mux.Handle("POST /todo/templates", requireAuth(handlers.HandleCreateTemplate(nc, handlerLogger)))
	mux.Handle("GET /todo/templates/{id}", requireAuth(handlers.HandleTemplateDetail(nc, handlerLogger)))
	mux.Handle("DELETE /todo/templates/{id}", requireAuth(handlers.HandleDeleteTemplate(nc, handlerLogger)))
	mux.Handle("POST /todo/templates/{id}/title", requireAuth(handlers.HandleUpdateTemplateTitle(nc, handlerLogger)))
	mux.Handle("POST /todo/templates/{id}/items", requireAuth(handlers.HandleCreateTemplateItem(nc, handlerLogger)))
	mux.Handle("DELETE /todo/templates/{id}/items/{itemId}", requireAuth(handlers.HandleDeleteTemplateItem(nc, handlerLogger)))
	mux.Handle("GET /todo/lists/{id}", requireAuth(handlers.HandleListDetail(nc, handlerLogger)))
	mux.Handle("GET /todo/lists/{id}/settings", requireAuth(handlers.HandleListSettings(nc, handlerLogger)))
	mux.Handle("DELETE /todo/lists/{id}", requireAuth(handlers.HandleDeleteList(nc, handlerLogger)))
	mux.Handle("POST /todo/lists/{id}/title", requireAuth(handlers.HandleUpdateListTitle(nc, handlerLogger)))
	mux.Handle("POST /todo/lists/{id}/items", requireAuth(handlers.HandleCreateItem(nc, sseManager, handlerLogger)))
	mux.Handle("POST /todo/items/{id}/toggle", requireAuth(handlers.HandleToggleItem(nc, sseManager, handlerLogger)))
	mux.Handle("DELETE /todo/items/{id}", requireAuth(handlers.HandleDeleteItem(nc, sseManager, handlerLogger)))
	mux.Handle("GET /todo/lists/{id}/events", requireAuth(handlers.HandleSSE(nc, sseManager, handlerLogger)))
	mux.Handle("POST /todo/lists/{id}/members", requireAuth(handlers.HandleAddListMember(nc, handlerLogger)))
	mux.Handle("DELETE /todo/lists/{id}/members/{memberID}", requireAuth(handlers.HandleRemoveListMember(nc, handlerLogger)))
	mux.Handle("POST /todo/lists/{id}/leave", requireAuth(handlers.HandleLeaveList(nc, handlerLogger)))

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
