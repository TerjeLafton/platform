package handlers

import (
	"log/slog"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	"github.com/terjelafton/platform/apps/web/internal/templates"
	applogger "github.com/terjelafton/platform/libs/logger"
)

func HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	templates.LoginPage("").Render(r.Context(), w)
}

func HandleLogin(nc *nats.Conn, cookieName string, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		token, err := natsclient.Login(nc, email, password, correlationID)
		if err != nil {
			logger.Warn("login failed", "email", email, "error", err)
			templates.LoginPage(err.Error()).Render(r.Context(), w)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token,
			Path:     "/",
			MaxAge:   7 * 24 * 60 * 60,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	templates.RegisterPage("").Render(r.Context(), w)
}

func HandleRegister(nc *nats.Conn, cookieName string, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		passwordConfirm := r.FormValue("password_confirm")

		if password != passwordConfirm {
			templates.RegisterPage("Passwords do not match").Render(r.Context(), w)
			return
		}

		correlationID := applogger.CorrelationIDFromContext(r.Context())

		token, err := natsclient.Register(nc, email, password, name, correlationID)
		if err != nil {
			logger.Warn("registration failed", "email", email, "error", err)
			templates.RegisterPage(err.Error()).Render(r.Context(), w)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token,
			Path:     "/",
			MaxAge:   7 * 24 * 60 * 60,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func HandleLogout(cookieName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   cookieName,
			MaxAge: -1,
			Path:   "/",
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
