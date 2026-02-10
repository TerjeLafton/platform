package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/middleware"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	"github.com/terjelafton/platform/apps/web/internal/templates"
	applogger "github.com/terjelafton/platform/libs/logger"
)

func HandleLogsPage(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		// Parse query params for filters
		serviceFilter := r.URL.Query().Get("service")
		levelFilter := r.URL.Query().Get("level")
		corrFilter := r.URL.Query().Get("correlation_id")

		page := 1
		if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
			page = p
		}

		limit := int32(50)
		offset := int32((page - 1) * int(limit))

		logs, total, err := natsclient.QueryLogs(nc, serviceFilter, levelFilter, corrFilter, limit, offset, correlationID)
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to query logs", "error", err, "user_id", userID)
			http.Error(w, "Failed to load logs", http.StatusInternalServerError)
			return
		}

		templates.LogsPage(userID, logs, total, page, serviceFilter, levelFilter, corrFilter).Render(r.Context(), w)
	}
}
