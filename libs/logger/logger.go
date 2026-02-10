package logger

import (
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"
)

// New creates a *slog.Logger that dual-writes to stdout (JSON) and NATS (protobuf).
// Use slog's InfoContext/WarnContext/ErrorContext methods to include the correlation
// ID from context automatically.
func New(nc *nats.Conn, service string) *slog.Logger {
	stdout := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	handler := &natsHandler{
		inner:   stdout,
		nc:      nc,
		service: service,
	}

	return slog.New(handler).With("service", service)
}
