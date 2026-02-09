package handlers

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/middleware"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	"github.com/terjelafton/platform/apps/web/internal/templates"
)

func HandleProfilePage(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		user, err := natsclient.GetUser(nc, userID)
		if err != nil {
			logger.Error("failed to get user", "error", err, "user_id", userID)
			http.Error(w, "Failed to load profile", http.StatusInternalServerError)
			return
		}

		templates.ProfilePage(userID, user.Name, user.Email, "").Render(r.Context(), w)
	}
}

func HandleUploadAvatar(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20+512) // 1MB + overhead

		file, header, err := r.FormFile("avatar")
		if err != nil {
			logger.Warn("failed to read avatar upload", "error", err, "user_id", userID)
			user, _ := natsclient.GetUser(nc, userID)
			name, email := "", ""
			if user != nil {
				name, email = user.Name, user.Email
			}
			templates.ProfilePage(userID, name, email, "Failed to read file").Render(r.Context(), w)
			return
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")

		data, err := io.ReadAll(file)
		if err != nil {
			logger.Warn("failed to read avatar data", "error", err, "user_id", userID)
			user, _ := natsclient.GetUser(nc, userID)
			name, email := "", ""
			if user != nil {
				name, email = user.Name, user.Email
			}
			templates.ProfilePage(userID, name, email, "Failed to read file").Render(r.Context(), w)
			return
		}

		if err := natsclient.UpdateAvatar(nc, userID, data, contentType); err != nil {
			logger.Warn("failed to update avatar", "error", err, "user_id", userID)
			user, _ := natsclient.GetUser(nc, userID)
			name, email := "", ""
			if user != nil {
				name, email = user.Name, user.Email
			}
			templates.ProfilePage(userID, name, email, err.Error()).Render(r.Context(), w)
			return
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}
