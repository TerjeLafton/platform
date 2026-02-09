package handlers

import (
	"log/slog"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/web/internal/middleware"
	"github.com/terjelafton/platform/apps/web/internal/natsclient"
	"github.com/terjelafton/platform/apps/web/internal/templates"
)

func HandleListsPage(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		lists, err := natsclient.GetListsByUser(nc, userID)
		if err != nil {
			logger.Error("failed to get lists", "error", err, "user_id", userID)
			http.Error(w, "Failed to load lists", http.StatusInternalServerError)
			return
		}

		templates.ListsPage(lists).Render(r.Context(), w)
	}
}

func HandleCreateList(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		title := r.FormValue("title")

		_, err := natsclient.CreateList(nc, userID, title)
		if err != nil {
			logger.Warn("failed to create list", "error", err, "user_id", userID, "title", title)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/todo", http.StatusSeeOther)
	}
}

func HandleDeleteList(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")

		err := natsclient.DeleteList(nc, id, userID)
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

		lists, err := natsclient.GetListsByUser(nc, userID)
		if err != nil {
			logger.Error("failed to get lists", "error", err, "user_id", userID)
			http.Error(w, "Failed to load list", http.StatusInternalServerError)
			return
		}

		// Find the specific list
		var found bool
		for _, list := range lists {
			if list.Id == id {
				items, err := natsclient.GetAllItemsFromList(nc, id, userID)
				if err != nil {
					logger.Error("failed to get items", "error", err, "user_id", userID, "list_id", id)
					http.Error(w, "Failed to load items", http.StatusInternalServerError)
					return
				}

				templates.ListDetailPage(list, items).Render(r.Context(), w)
				found = true
				break
			}
		}

		if !found {
			http.Redirect(w, r, "/todo", http.StatusSeeOther)
		}
	}
}

func HandleUpdateListTitle(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")
		title := r.FormValue("title")

		list, err := natsclient.UpdateListTitle(nc, id, userID, title)
		if err != nil {
			logger.Warn("failed to update list title", "error", err, "user_id", userID, "list_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		templates.ListTitle(list).Render(r.Context(), w)
	}
}

func HandleCreateItem(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		listID := r.PathValue("id")
		title := r.FormValue("title")

		item, err := natsclient.CreateItem(nc, listID, userID, title)
		if err != nil {
			logger.Warn("failed to create item", "error", err, "user_id", userID, "list_id", listID)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		templates.ItemRow(item).Render(r.Context(), w)
	}
}

func HandleToggleItem(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")

		item, err := natsclient.ToggleItemCompleted(nc, id, userID)
		if err != nil {
			logger.Warn("failed to toggle item", "error", err, "user_id", userID, "item_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		templates.ItemRow(item).Render(r.Context(), w)
	}
}

func HandleDeleteItem(nc *nats.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		id := r.PathValue("id")

		err := natsclient.DeleteItem(nc, id, userID)
		if err != nil {
			logger.Warn("failed to delete item", "error", err, "user_id", userID, "item_id", id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
