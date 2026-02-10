package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/middleware"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	"github.com/terjelafton/platform/apps/web/internal/sse"
	"github.com/terjelafton/platform/apps/web/internal/templates"
	applogger "github.com/terjelafton/platform/libs/logger"
)

func HandleListsPage(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		lists, err := natsclient.GetListsByUser(nc, userID, correlationID)
		if err != nil {
			logger.Error("failed to get lists", "error", err, "user_id", userID)
			http.Error(w, "Failed to load lists", http.StatusInternalServerError)
			return
		}

		templates.ListsPage(userID, lists).Render(r.Context(), w)
	}
}

func HandleDeleteList(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		err := natsclient.DeleteList(nc, id, userID, correlationID)
		if err != nil {
			logger.Warn("failed to delete list", "error", err, "user_id", userID, "list_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleListDetail(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		lists, err := natsclient.GetListsByUser(nc, userID, correlationID)
		if err != nil {
			logger.Error("failed to get lists", "error", err, "user_id", userID)
			http.Error(w, "Failed to load list", http.StatusInternalServerError)
			return
		}

		var found bool
		for _, list := range lists {
			if list.Id == id {
				items, err := natsclient.GetAllItemsFromList(nc, id, userID, correlationID)
				if err != nil {
					logger.Error("failed to get items", "error", err, "user_id", userID, "list_id", id)
					http.Error(w, "Failed to load items", http.StatusInternalServerError)
					return
				}

				templates.ListDetailPage(userID, list, items).Render(r.Context(), w)
				found = true
				break
			}
		}

		if !found {
			http.Redirect(w, r, "/todo", http.StatusSeeOther)
		}
	}
}

func HandleListSettings(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		lists, err := natsclient.GetListsByUser(nc, userID, correlationID)
		if err != nil {
			logger.Error("failed to get lists", "error", err, "user_id", userID)
			http.Error(w, "Failed to load list", http.StatusInternalServerError)
			return
		}

		for _, list := range lists {
			if list.Id == id {
				if !list.IsOwner {
					http.Redirect(w, r, fmt.Sprintf("/todo/lists/%s", id), http.StatusSeeOther)
					return
				}

				members, err := natsclient.GetListMembers(nc, id, userID, correlationID)
				if err != nil {
					logger.Error("failed to get members", "error", err, "user_id", userID, "list_id", id)
					http.Error(w, "Failed to load members", http.StatusInternalServerError)
					return
				}

				templates.ListSettingsPage(userID, list, members).Render(r.Context(), w)
				return
			}
		}

		http.Redirect(w, r, "/todo", http.StatusSeeOther)
	}
}

func HandleUpdateListTitle(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")
		title := r.FormValue("title")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		_, err := natsclient.UpdateListTitle(nc, id, userID, title, correlationID)
		if err != nil {
			logger.Warn("failed to update list title", "error", err, "user_id", userID, "list_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/todo/lists/%s/settings", id), http.StatusSeeOther)
	}
}

func HandleCreateItem(nc *nats.Conn, sseManager *sse.Manager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		listID := r.PathValue("id")
		title := r.FormValue("title")

		correlationID := applogger.CorrelationIDFromContext(r.Context())
		sseManager.AddCorrelationID(userID, correlationID)

		item, err := natsclient.CreateItem(nc, listID, userID, title, correlationID)
		if err != nil {
			logger.Warn("failed to create item", "error", err, "user_id", userID, "list_id", listID)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		templates.ItemRow(item).Render(r.Context(), w)
	}
}

func HandleToggleItem(nc *nats.Conn, sseManager *sse.Manager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")

		correlationID := applogger.CorrelationIDFromContext(r.Context())
		sseManager.AddCorrelationID(userID, correlationID)

		item, err := natsclient.ToggleItemCompleted(nc, id, userID, correlationID)
		if err != nil {
			logger.Warn("failed to toggle item", "error", err, "user_id", userID, "item_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		templates.ItemRow(item).Render(r.Context(), w)
	}
}

func HandleDeleteItem(nc *nats.Conn, sseManager *sse.Manager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")
		listID := r.URL.Query().Get("list_id")

		correlationID := applogger.CorrelationIDFromContext(r.Context())
		sseManager.AddCorrelationID(userID, correlationID)

		err := natsclient.DeleteItem(nc, id, userID, listID, correlationID)
		if err != nil {
			logger.Warn("failed to delete item", "error", err, "user_id", userID, "item_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleAddListMember(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		listID := r.PathValue("id")
		email := r.FormValue("email")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		member, err := natsclient.AddListMember(nc, listID, userID, email, correlationID)
		if err != nil {
			logger.Warn("failed to add member", "error", err, "user_id", userID, "list_id", listID, "email", email)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		templates.MemberRow(listID, member).Render(r.Context(), w)
	}
}

func HandleRemoveListMember(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		listID := r.PathValue("id")
		memberID := r.PathValue("memberID")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		err := natsclient.RemoveListMember(nc, listID, userID, memberID, correlationID)
		if err != nil {
			logger.Warn("failed to remove member", "error", err, "user_id", userID, "list_id", listID, "member_id", memberID)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleLeaveList(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		listID := r.PathValue("id")
		correlationID := applogger.CorrelationIDFromContext(r.Context())

		err := natsclient.RemoveListMember(nc, listID, userID, userID, correlationID)
		if err != nil {
			logger.Warn("failed to leave list", "error", err, "user_id", userID, "list_id", listID)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
