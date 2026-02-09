package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/middleware"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	"github.com/terjelafton/platform/apps/web/internal/templates"
)

func HandleCreateListPage(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		templateList, err := natsclient.GetTemplatesByUser(nc, userID)
		if err != nil {
			logger.Error("failed to get templates", "error", err, "user_id", userID)
			http.Error(w, "Failed to load templates", http.StatusInternalServerError)
			return
		}

		templates.CreateListPage(userID, templateList).Render(r.Context(), w)
	}
}

func HandleCreateListFromForm(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		title := r.FormValue("title")
		templateID := r.FormValue("template_id")

		if templateID != "" {
			list, err := natsclient.UseTemplate(nc, templateID, userID, title)
			if err != nil {
				logger.Warn("failed to use template", "error", err, "user_id", userID, "template_id", templateID)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, fmt.Sprintf("/todo/%s", list.Id), http.StatusSeeOther)
			return
		}

		if title == "" {
			http.Error(w, "Title is required when not using a template", http.StatusBadRequest)
			return
		}

		list, err := natsclient.CreateList(nc, userID, title)
		if err != nil {
			logger.Warn("failed to create list", "error", err, "user_id", userID, "title", title)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/todo/%s", list.Id), http.StatusSeeOther)
	}
}

func HandleTemplatesPage(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		templateList, err := natsclient.GetTemplatesByUser(nc, userID)
		if err != nil {
			logger.Error("failed to get templates", "error", err, "user_id", userID)
			http.Error(w, "Failed to load templates", http.StatusInternalServerError)
			return
		}

		templates.TemplatesPage(userID, templateList).Render(r.Context(), w)
	}
}

func HandleCreateTemplate(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		title := r.FormValue("title")

		_, err := natsclient.CreateTemplate(nc, userID, title)
		if err != nil {
			logger.Warn("failed to create template", "error", err, "user_id", userID, "title", title)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/todo/templates", http.StatusSeeOther)
	}
}

func HandleDeleteTemplate(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")

		err := natsclient.DeleteTemplate(nc, id, userID)
		if err != nil {
			logger.Warn("failed to delete template", "error", err, "user_id", userID, "template_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleTemplateDetail(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")

		templateList, err := natsclient.GetTemplatesByUser(nc, userID)
		if err != nil {
			logger.Error("failed to get templates", "error", err, "user_id", userID)
			http.Error(w, "Failed to load templates", http.StatusInternalServerError)
			return
		}

		for _, t := range templateList {
			if t.Id == id {
				items, err := natsclient.GetTemplateItems(nc, id, userID)
				if err != nil {
					logger.Error("failed to get template items", "error", err, "user_id", userID, "template_id", id)
					http.Error(w, "Failed to load template items", http.StatusInternalServerError)
					return
				}

				templates.TemplateDetailPage(userID, t, items).Render(r.Context(), w)
				return
			}
		}

		http.Redirect(w, r, "/todo/templates", http.StatusSeeOther)
	}
}

func HandleUpdateTemplateTitle(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")
		title := r.FormValue("title")

		_, err := natsclient.UpdateTemplateTitle(nc, id, userID, title)
		if err != nil {
			logger.Warn("failed to update template title", "error", err, "user_id", userID, "template_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/todo/templates/%s", id), http.StatusSeeOther)
	}
}

func HandleCreateTemplateItem(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		templateID := r.PathValue("id")
		title := r.FormValue("title")

		item, err := natsclient.CreateTemplateItem(nc, templateID, userID, title)
		if err != nil {
			logger.Warn("failed to create template item", "error", err, "user_id", userID, "template_id", templateID)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		templates.TemplateItemRow(templateID, item).Render(r.Context(), w)
	}
}

func HandleDeleteTemplateItem(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("itemId")

		err := natsclient.DeleteTemplateItem(nc, id, userID)
		if err != nil {
			logger.Warn("failed to delete template item", "error", err, "user_id", userID, "template_item_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
