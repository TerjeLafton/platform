package nats

import "github.com/nats-io/nats.go"

func (h *Handler) Register(nc *nats.Conn) error {
	subjects := map[string]nats.MsgHandler{
		"id.auth.register": h.HandleRegister,
		"id.auth.login":    h.HandleLogin,
		"id.auth.validate": h.HandleValidateToken,
	}

	for subject, handler := range subjects {
		if _, err := nc.Subscribe(subject, handler); err != nil {
			return err
		}
		h.logger.Info("subscribed to subject", "subject", subject)
	}

	return nil
}
