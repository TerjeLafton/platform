package nats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	logv1 "github.com/terjelafton/platform/libs/proto-stubs/log/v1"
	"google.golang.org/protobuf/proto"
)

func (h *Handler) HandleIngest(msg *nats.Msg) {
	var record logv1.LogRecord
	if err := proto.Unmarshal(msg.Data, &record); err != nil {
		h.logger.Warn("failed to unmarshal log record", "error", err)
		return
	}

	timestamp, err := time.Parse(time.RFC3339Nano, record.Timestamp)
	if err != nil {
		h.logger.Warn("invalid timestamp in log record", "error", err, "timestamp", record.Timestamp)
		return
	}

	var attrs json.RawMessage
	if len(record.Attrs) > 0 {
		attrs = json.RawMessage(record.Attrs)
	}

	if err := h.service.IngestLog(
		context.Background(),
		timestamp,
		record.Level,
		record.Service,
		record.Module,
		record.CorrelationId,
		record.Message,
		attrs,
	); err != nil {
		h.logger.Error("failed to ingest log", "error", err)
	}
}
