package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	applogger "github.com/terjelafton/platform/libs/logger"
)

type contextKey struct{}

func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKey{}).(string); ok {
		return v
	}
	return ""
}

func RequireAuth(nc *nats.Conn, cookieName string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			correlationID := applogger.CorrelationIDFromContext(r.Context())

			userID, err := natsclient.ValidateToken(nc, cookie.Value, correlationID)
			if err != nil {
				if _, ok := err.(*natsclient.ServiceError); ok {
					logger.Info("invalid token, redirecting to login")
				} else {
					logger.Error("failed to validate token", "error", err)
				}

				http.SetCookie(w, &http.Cookie{
					Name:   cookieName,
					MaxAge: -1,
					Path:   "/",
				})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), contextKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
