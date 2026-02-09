package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/terjelafton/platform/apps/web/internal/middleware"
	"github.com/terjelafton/platform/apps/web/internal/sse"
	"github.com/terjelafton/platform/apps/web/internal/templates"
	todov1 "github.com/terjelafton/platform/libs/proto-stubs/todo/v1"
)

func HandleSSE(nc *nats.Conn, manager *sse.Manager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		userID := middleware.UserIDFromContext(r.Context())
		listID := r.PathValue("id")

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Create and register SSE connection
		connID := uuid.New().String()
		conn := sse.NewConn(connID, userID, listID)
		manager.Register(conn)
		defer manager.Unregister(conn)

		logger.Info("SSE connection opened", "conn_id", connID, "user_id", userID, "list_id", listID)

		// Subscribe to NATS events for this list
		subject := fmt.Sprintf("todo.events.%s.>", listID)
		sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
			correlationID := msg.Header.Get("X-Correlation-ID")
			if correlationID != "" && conn.HasCorrelationID(correlationID) {
				return
			}

			// Extract event type from subject: todo.events.<list_id>.item.created → item.created
			parts := strings.SplitN(msg.Subject, ".", 4)
			if len(parts) < 4 {
				return
			}
			eventType := parts[3]

			html, err := renderEventHTML(r.Context(), eventType, msg.Data)
			if err != nil {
				logger.Warn("failed to render event", "error", err, "event_type", eventType)
				return
			}

			select {
			case conn.Events <- sse.Event{Name: eventType, Data: html}:
			default:
				logger.Warn("SSE event channel full, dropping event", "conn_id", connID)
			}
		})
		if err != nil {
			logger.Error("failed to subscribe to NATS", "error", err, "subject", subject)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer sub.Unsubscribe()

		// Stream events to the client
		for {
			select {
			case event := <-conn.Events:
				fmt.Fprintf(w, "event: %s\n", event.Name)
				for _, line := range strings.Split(event.Data, "\n") {
					fmt.Fprintf(w, "data: %s\n", line)
				}
				fmt.Fprint(w, "\n")
				flusher.Flush()
			case <-r.Context().Done():
				logger.Info("SSE connection closed", "conn_id", connID, "user_id", userID, "list_id", listID)
				return
			}
		}
	}
}

func renderEventHTML(ctx context.Context, eventType string, data []byte) (string, error) {
	switch eventType {
	case "item.created":
		var item todov1.Item
		if err := proto.Unmarshal(data, &item); err != nil {
			return "", fmt.Errorf("unmarshal item: %w", err)
		}
		var buf bytes.Buffer
		if err := templates.ItemRow(&item).Render(ctx, &buf); err != nil {
			return "", fmt.Errorf("render item row: %w", err)
		}
		return buf.String(), nil

	case "item.toggled":
		var item todov1.Item
		if err := proto.Unmarshal(data, &item); err != nil {
			return "", fmt.Errorf("unmarshal item: %w", err)
		}
		var buf bytes.Buffer
		if err := templates.ItemRowOOB(&item).Render(ctx, &buf); err != nil {
			return "", fmt.Errorf("render item row oob: %w", err)
		}
		return buf.String(), nil

	case "item.deleted":
		var event todov1.ItemDeletedEvent
		if err := proto.Unmarshal(data, &event); err != nil {
			return "", fmt.Errorf("unmarshal delete event: %w", err)
		}
		return fmt.Sprintf(`<div id="item-%s" hx-swap-oob="delete"></div>`, event.ItemId), nil

	default:
		return "", fmt.Errorf("unknown event type: %s", eventType)
	}
}
