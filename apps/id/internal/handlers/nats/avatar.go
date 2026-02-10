package nats

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/id/internal/service"
	applogger "github.com/terjelafton/platform/libs/logger"
	idv1 "github.com/terjelafton/platform/libs/proto-stubs/id/v1"
	"google.golang.org/protobuf/proto"
)

func (h *Handler) HandleGetAvatar(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req idv1.GetAvatarRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal get avatar request", "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user ID", "user_id", req.UserId)
		h.respondError(msg, "VALIDATION_ERROR", "Invalid user ID")
		return
	}

	data, contentType, err := h.service.GetAvatar(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get avatar", "error", err, "user_id", req.UserId)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	resp := &idv1.GetAvatarResponse{
		Avatar:      data,
		ContentType: contentType,
	}
	h.respondSuccess(msg, resp)
}

func (h *Handler) HandleUpdateAvatar(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req idv1.UpdateAvatarRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal update avatar request", "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user ID", "user_id", req.UserId)
		h.respondError(msg, "VALIDATION_ERROR", "Invalid user ID")
		return
	}

	if err := h.service.UpdateAvatar(ctx, userID, req.Avatar, req.ContentType); err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "avatar validation failed", "field", valErr.Field, "error", valErr.Message)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Error())
			return
		}
		h.logger.ErrorContext(ctx, "avatar update failed", "error", err)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	h.respondSuccess(msg, &idv1.UpdateAvatarResponse{})
}
