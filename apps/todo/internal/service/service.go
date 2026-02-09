package service

import (
	"log/slog"

	"github.com/nats-io/nats.go"

	"github.com/terjelafton/platform/apps/todo/internal/db"
)

type Service struct {
	queries *db.Queries
	nc      *nats.Conn
	logger  *slog.Logger
}

func New(queries *db.Queries, nc *nats.Conn, logger *slog.Logger) *Service {
	return &Service{
		queries: queries,
		nc:      nc,
		logger:  logger.With("module", "service"),
	}
}
