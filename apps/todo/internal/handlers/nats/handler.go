package nats

import (
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/todo/internal/service"
	commonv1 "github.com/terjelafton/platform/libs/proto-stubs/common/v1"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	service *service.Service
	logger  *slog.Logger
}

func New(svc *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With("component", "nats-handler"),
	}
}

func (h *Handler) respondError(msg *nats.Msg, code, message string) {
	errResp := &commonv1.ErrorResponse{
		Code:    code,
		Message: message,
	}
	data, _ := proto.Marshal(errResp)
	msg.Respond(data)
}

func (h *Handler) respondSuccess(msg *nats.Msg, resp proto.Message) {
	data, err := proto.Marshal(resp)
	if err != nil {
		h.logger.Error("failed to marshal response", "error", err)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}
	msg.Respond(data)
}
