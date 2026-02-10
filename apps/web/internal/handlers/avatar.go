package handlers

import (
	"log/slog"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	applogger "github.com/terjelafton/platform/libs/logger"
)

func HandleAvatar(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		data, contentType, err := natsclient.GetAvatar(nc, userID, correlationID)
		if err != nil {
			logger.WarnContext(r.Context(), "failed to get avatar", "error", err, "user_id", userID)
			http.Redirect(w, r, "/static/default-avatar.svg", http.StatusSeeOther)
			return
		}

		if data == nil {
			http.Redirect(w, r, "/static/default-avatar.svg", http.StatusSeeOther)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(data)
	}
}
