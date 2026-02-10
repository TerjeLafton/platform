package natsclient

import (
	"fmt"

	"github.com/nats-io/nats.go"
	logv1 "github.com/terjelafton/platform/libs/proto-stubs/log/v1"
	"google.golang.org/protobuf/proto"
)

func QueryLogs(nc *nats.Conn, service, level, correlationID string, limit, offset int32, reqCorrelationID string) ([]*logv1.LogRecord, int32, error) {
	req := &logv1.QueryLogsRequest{
		Service:       service,
		Level:         level,
		CorrelationId: correlationID,
		Limit:         limit,
		Offset:        offset,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal query logs request: %w", err)
	}

	msg, err := request(nc, "log.query", data, reqCorrelationID)
	if err != nil {
		return nil, 0, fmt.Errorf("nats request: %w", err)
	}

	var resp logv1.QueryLogsResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil {
		return resp.Logs, resp.Total, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, 0, svcErr
	}

	return nil, 0, fmt.Errorf("unexpected response from log service")
}
