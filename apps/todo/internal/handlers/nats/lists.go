package nats

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/terjelafton/platform/apps/todo/internal/service"
	applogger "github.com/terjelafton/platform/libs/logger"
	todov1 "github.com/terjelafton/platform/libs/proto-stubs/todo/v1"
)

func (h *Handler) HandleCreateList(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.CreateListRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	list, err := h.service.CreateList(ctx, userID, req.Title)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "validation failed", "error", err, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.CreateListResponse{
		List: &todov1.List{
			Id:        list.ID.String(),
			UserId:    list.UserID.String(),
			Title:     list.Title,
			CreatedAt: timestamppb.New(list.CreatedAt),
			UpdatedAt: timestamppb.New(list.UpdatedAt),
		},
	})
}

func (h *Handler) HandleGetListsByUser(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.GetListsByUserRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	lists, err := h.service.GetListsByUser(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "service error", "error", err, "user_id", userID)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	pbLists := make([]*todov1.List, len(lists))
	for i, list := range lists {
		pbLists[i] = &todov1.List{
			Id:             list.ID.String(),
			UserId:         list.UserID.String(),
			Title:          list.Title,
			CreatedAt:      timestamppb.New(list.CreatedAt),
			UpdatedAt:      timestamppb.New(list.UpdatedAt),
			TotalItems:     list.TotalItems,
			CompletedItems: list.CompletedItems,
			OwnerName:      list.OwnerName,
			IsOwner:        list.IsOwner,
		}
	}

	h.respondSuccess(msg, &todov1.GetListsByUserResponse{
		Lists: pbLists,
	})
}

func (h *Handler) HandleUpdateListTitle(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.UpdateListTitleRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	listID, err := uuid.Parse(req.Id)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid list_id", "subject", msg.Subject, "list_id", req.Id)
		h.respondError(msg, "INVALID_REQUEST", "Invalid list ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	list, err := h.service.UpdateListTitle(ctx, listID, userID, req.Title)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "validation failed", "error", err, "list_id", listID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "list_id", listID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.UpdateListTitleResponse{
		List: &todov1.List{
			Id:        list.ID.String(),
			UserId:    list.UserID.String(),
			Title:     list.Title,
			CreatedAt: timestamppb.New(list.CreatedAt),
			UpdatedAt: timestamppb.New(list.UpdatedAt),
		},
	})
}

func (h *Handler) HandleDeleteList(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.DeleteListRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	listID, err := uuid.Parse(req.Id)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid list_id", "subject", msg.Subject, "list_id", req.Id)
		h.respondError(msg, "INVALID_REQUEST", "Invalid list ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	if err := h.service.DeleteList(ctx, listID, userID); err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "delete failed", "error", valErr.Message, "list_id", listID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "list_id", listID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.DeleteListResponse{})
}
