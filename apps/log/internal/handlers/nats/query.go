package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	logv1 "github.com/terjelafton/platform/libs/proto-stubs/log/v1"
	"google.golang.org/protobuf/proto"
)

func (h *Handler) HandleQueryLogs(msg *nats.Msg) {
	var req logv1.QueryLogsRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Warn("failed to unmarshal query request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	entries, total, err := h.service.QueryLogs(
		context.Background(),
		req.Service,
		req.Level,
		req.CorrelationId,
		req.Limit,
		req.Offset,
	)
	if err != nil {
		h.logger.Error("service error", "error", err)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	logs := make([]*logv1.LogRecord, len(entries))
	for i, e := range entries {
		logs[i] = &logv1.LogRecord{
			Timestamp:     e.Timestamp.Format("2006-01-02T15:04:05.000Z07:00"),
			Level:         e.Level,
			Service:       e.Service,
			Module:        e.Module,
			CorrelationId: e.CorrelationID,
			Message:       e.Message,
			Attrs:         e.Attrs,
		}
	}

	h.respondSuccess(msg, &logv1.QueryLogsResponse{
		Logs:  logs,
		Total: int32(total),
	})
}
