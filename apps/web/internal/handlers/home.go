package handlers

import (
	"log/slog"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/middleware"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	"github.com/terjelafton/platform/apps/web/internal/templates"
)

func HandleHomePage(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		user, err := natsclient.GetUser(nc, userID)
		if err != nil {
			logger.Error("failed to get user", "error", err, "user_id", userID)
			http.Error(w, "Failed to load home", http.StatusInternalServerError)
			return
		}

		templates.HomePage(userID, user.Name).Render(r.Context(), w)
	}
}
