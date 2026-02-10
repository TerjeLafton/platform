package service

import (
	"log/slog"

	"github.com/terjelafton/platform/apps/log/internal/db"
)

type Service struct {
	queries *db.Queries
	logger  *slog.Logger
}

func New(queries *db.Queries, logger *slog.Logger) *Service {
	return &Service{
		queries: queries,
		logger:  logger.With("module", "service"),
	}
}
