package service

import (
	"log/slog"

	"github.com/terjelafton/platform/apps/id/internal/db"
)

type Service struct {
	queries   db.Querier
	jwtSecret string
	logger    *slog.Logger
}

func New(queries db.Querier, jwtSecret string, logger *slog.Logger) *Service {
	return &Service{
		queries:   queries,
		jwtSecret: jwtSecret,
		logger:    logger.With("module", "service"),
	}
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
