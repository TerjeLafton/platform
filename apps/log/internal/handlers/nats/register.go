package nats

import "github.com/nats-io/nats.go"

func (h *Handler) Register(nc *nats.Conn) error {
	// Fire-and-forget subscription (no reply)
	if _, err := nc.Subscribe("log.ingest", h.HandleIngest); err != nil {
		return err
	}
	h.logger.Info("subscribed to subject", "subject", "log.ingest")

	// Request/reply subscriptions
	subjects := map[string]nats.MsgHandler{
		"log.query": h.HandleQueryLogs,
	}

	for subject, handler := range subjects {
		if _, err := nc.Subscribe(subject, handler); err != nil {
			return err
		}
		h.logger.Info("subscribed to subject", "subject", subject)
	}

	return nil
}
