package logger

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"time"

	"github.com/nats-io/nats.go"
	logv1 "github.com/terjelafton/platform/libs/proto-stubs/log/v1"
	"google.golang.org/protobuf/proto"
)

// natsHandler is a slog.Handler that dual-writes: JSON to an inner handler
// (stdout) and protobuf to NATS on the "log.ingest" subject.
type natsHandler struct {
	inner   slog.Handler
	nc      *nats.Conn
	service string
	attrs   []slog.Attr
}

func (h *natsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *natsHandler) Handle(ctx context.Context, r slog.Record) error {
	// Add correlation ID from context if present
	correlationID := CorrelationIDFromContext(ctx)
	if correlationID != "" {
		r.AddAttrs(slog.String("correlation_id", correlationID))
	}

	// Always write to stdout
	if err := h.inner.Handle(ctx, r); err != nil {
		return err
	}

	// Build log record for NATS
	record := &logv1.LogRecord{
		Timestamp: r.Time.Format(time.RFC3339Nano),
		Level:     r.Level.String(),
		Service:   h.service,
		Message:   r.Message,
	}

	// Extract known fields, collect rest into attrs map
	extra := make(map[string]string)

	extractAttr := func(a slog.Attr) {
		switch a.Key {
		case "service":
			// Already set from constructor
		case "module":
			record.Module = a.Value.String()
		case "correlation_id":
			record.CorrelationId = a.Value.String()
		default:
			extra[a.Key] = a.Value.String()
		}
	}

	for _, a := range h.attrs {
		extractAttr(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		extractAttr(a)
		return true
	})

	if len(extra) > 0 {
		jsonBytes, _ := json.Marshal(extra)
		record.Attrs = jsonBytes
	}

	// Fire-and-forget publish to NATS
	data, err := proto.Marshal(record)
	if err != nil {
		return nil // Don't fail the log call
	}
	h.nc.Publish("log.ingest", data)

	return nil
}

func (h *natsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &natsHandler{
		inner:   h.inner.WithAttrs(attrs),
		nc:      h.nc,
		service: h.service,
		attrs:   append(slices.Clone(h.attrs), attrs...),
	}
}

func (h *natsHandler) WithGroup(name string) slog.Handler {
	return &natsHandler{
		inner:   h.inner.WithGroup(name),
		nc:      h.nc,
		service: h.service,
		attrs:   h.attrs,
	}
}
