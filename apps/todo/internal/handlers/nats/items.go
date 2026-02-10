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

func (h *Handler) HandleCreateItem(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.CreateItemRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	listID, err := uuid.Parse(req.ListId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid list_id", "subject", msg.Subject, "list_id", req.ListId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid list ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	item, err := h.service.CreateItem(ctx, listID, userID, req.Title, correlationID)
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

	h.respondSuccess(msg, &todov1.CreateItemResponse{
		Item: &todov1.Item{
			Id:        item.ID.String(),
			ListId:    item.ListID.String(),
			Title:     item.Title,
			Completed: item.Completed,
			CreatedAt: timestamppb.New(item.CreatedAt),
			UpdatedAt: timestamppb.New(item.UpdatedAt),
		},
	})
}

func (h *Handler) HandleGetAllItemsFromList(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.GetAllItemsFromListRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	listID, err := uuid.Parse(req.ListId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid list_id", "subject", msg.Subject, "list_id", req.ListId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid list ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	items, err := h.service.GetAllItemsFromList(ctx, listID, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "service error", "error", err, "list_id", listID, "user_id", userID)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	pbItems := make([]*todov1.Item, len(items))
	for i, item := range items {
		pbItems[i] = &todov1.Item{
			Id:        item.ID.String(),
			ListId:    item.ListID.String(),
			Title:     item.Title,
			Completed: item.Completed,
			CreatedAt: timestamppb.New(item.CreatedAt),
			UpdatedAt: timestamppb.New(item.UpdatedAt),
		}
	}

	h.respondSuccess(msg, &todov1.GetAllItemsFromListResponse{
		Items: pbItems,
	})
}

func (h *Handler) HandleToggleItemCompleted(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.ToggleItemCompletedRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	itemID, err := uuid.Parse(req.Id)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid item_id", "subject", msg.Subject, "item_id", req.Id)
		h.respondError(msg, "INVALID_REQUEST", "Invalid item ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	item, err := h.service.ToggleItemCompleted(ctx, itemID, userID, correlationID)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "toggle failed", "error", valErr.Message, "item_id", itemID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "item_id", itemID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.ToggleItemCompletedResponse{
		Item: &todov1.Item{
			Id:        item.ID.String(),
			ListId:    item.ListID.String(),
			Title:     item.Title,
			Completed: item.Completed,
			CreatedAt: timestamppb.New(item.CreatedAt),
			UpdatedAt: timestamppb.New(item.UpdatedAt),
		},
	})
}

func (h *Handler) HandleDeleteItem(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.DeleteItemRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	itemID, err := uuid.Parse(req.Id)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid item_id", "subject", msg.Subject, "item_id", req.Id)
		h.respondError(msg, "INVALID_REQUEST", "Invalid item ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	listID, err := uuid.Parse(req.ListId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid list_id", "subject", msg.Subject, "list_id", req.ListId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid list ID format")
		return
	}

	if err := h.service.DeleteItem(ctx, itemID, userID, listID, correlationID); err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "delete failed", "error", valErr.Message, "item_id", itemID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "item_id", itemID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.DeleteItemResponse{})
}
