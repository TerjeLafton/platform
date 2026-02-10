package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/terjelafton/platform/apps/log/internal/db"
)

func (s *Service) IngestLog(ctx context.Context, timestamp time.Time, level, svc, module, correlationID, message string, attrs json.RawMessage) error {
	if attrs == nil {
		attrs = json.RawMessage("{}")
	}

	err := s.queries.InsertLog(ctx, db.InsertLogParams{
		Timestamp:     timestamp,
		Level:         level,
		Service:       svc,
		Module:        module,
		CorrelationID: correlationID,
		Message:       message,
		Attrs:         attrs,
	})
	if err != nil {
		s.logger.Error("failed to insert log", "error", err)
		return err
	}

	return nil
}

func (s *Service) QueryLogs(ctx context.Context, svc, level, correlationID string, limit, offset int32) ([]db.LogEntry, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	entries, err := s.queries.QueryLogs(ctx, db.QueryLogsParams{
		Column1: svc,
		Column2: level,
		Column3: correlationID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		s.logger.Error("failed to query logs", "error", err)
		return nil, 0, err
	}

	count, err := s.queries.CountLogs(ctx, db.CountLogsParams{
		Column1: svc,
		Column2: level,
		Column3: correlationID,
	})
	if err != nil {
		s.logger.Error("failed to count logs", "error", err)
		return nil, 0, err
	}

	return entries, count, nil
}

func (s *Service) DeleteOldLogs(ctx context.Context, before time.Time) error {
	err := s.queries.DeleteOldLogs(ctx, before)
	if err != nil {
		s.logger.Error("failed to delete old logs", "error", err)
		return err
	}
	s.logger.Info("deleted old logs", "before", before)
	return nil
}
