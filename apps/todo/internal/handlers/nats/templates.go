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

func (h *Handler) HandleCreateTemplate(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.CreateTemplateRequest
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

	template, err := h.service.CreateTemplate(ctx, userID, req.Title)
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

	h.respondSuccess(msg, &todov1.CreateTemplateResponse{
		Template: &todov1.Template{
			Id:        template.ID.String(),
			UserId:    template.UserID.String(),
			Title:     template.Title,
			CreatedAt: timestamppb.New(template.CreatedAt),
			UpdatedAt: timestamppb.New(template.UpdatedAt),
		},
	})
}

func (h *Handler) HandleGetTemplatesByUser(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.GetTemplatesByUserRequest
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

	templates, err := h.service.GetTemplatesByUser(ctx, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "service error", "error", err, "user_id", userID)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	pbTemplates := make([]*todov1.Template, len(templates))
	for i, t := range templates {
		pbTemplates[i] = &todov1.Template{
			Id:        t.ID.String(),
			UserId:    t.UserID.String(),
			Title:     t.Title,
			CreatedAt: timestamppb.New(t.CreatedAt),
			UpdatedAt: timestamppb.New(t.UpdatedAt),
			ItemCount: t.ItemCount,
		}
	}

	h.respondSuccess(msg, &todov1.GetTemplatesByUserResponse{
		Templates: pbTemplates,
	})
}

func (h *Handler) HandleUpdateTemplateTitle(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.UpdateTemplateTitleRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	templateID, err := uuid.Parse(req.Id)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid template_id", "subject", msg.Subject, "template_id", req.Id)
		h.respondError(msg, "INVALID_REQUEST", "Invalid template ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	template, err := h.service.UpdateTemplateTitle(ctx, templateID, userID, req.Title)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "validation failed", "error", err, "template_id", templateID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "template_id", templateID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.UpdateTemplateTitleResponse{
		Template: &todov1.Template{
			Id:        template.ID.String(),
			UserId:    template.UserID.String(),
			Title:     template.Title,
			CreatedAt: timestamppb.New(template.CreatedAt),
			UpdatedAt: timestamppb.New(template.UpdatedAt),
		},
	})
}

func (h *Handler) HandleDeleteTemplate(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.DeleteTemplateRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	templateID, err := uuid.Parse(req.Id)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid template_id", "subject", msg.Subject, "template_id", req.Id)
		h.respondError(msg, "INVALID_REQUEST", "Invalid template ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	if err := h.service.DeleteTemplate(ctx, templateID, userID); err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "delete failed", "error", valErr.Message, "template_id", templateID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "template_id", templateID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.DeleteTemplateResponse{})
}

func (h *Handler) HandleCreateTemplateItem(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.CreateTemplateItemRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	templateID, err := uuid.Parse(req.TemplateId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid template_id", "subject", msg.Subject, "template_id", req.TemplateId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid template ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	item, err := h.service.CreateTemplateItem(ctx, templateID, userID, req.Title)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "validation failed", "error", err, "template_id", templateID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "template_id", templateID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.CreateTemplateItemResponse{
		Item: &todov1.TemplateItem{
			Id:         item.ID.String(),
			TemplateId: item.TemplateID.String(),
			Title:      item.Title,
			CreatedAt:  timestamppb.New(item.CreatedAt),
		},
	})
}

func (h *Handler) HandleGetTemplateItems(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.GetTemplateItemsRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	templateID, err := uuid.Parse(req.TemplateId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid template_id", "subject", msg.Subject, "template_id", req.TemplateId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid template ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	items, err := h.service.GetTemplateItems(ctx, templateID, userID)
	if err != nil {
		h.logger.ErrorContext(ctx, "service error", "error", err, "template_id", templateID, "user_id", userID)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	pbItems := make([]*todov1.TemplateItem, len(items))
	for i, item := range items {
		pbItems[i] = &todov1.TemplateItem{
			Id:         item.ID.String(),
			TemplateId: item.TemplateID.String(),
			Title:      item.Title,
			CreatedAt:  timestamppb.New(item.CreatedAt),
		}
	}

	h.respondSuccess(msg, &todov1.GetTemplateItemsResponse{
		Items: pbItems,
	})
}

func (h *Handler) HandleDeleteTemplateItem(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.DeleteTemplateItemRequest
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

	if err := h.service.DeleteTemplateItem(ctx, itemID, userID); err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "delete failed", "error", valErr.Message, "item_id", itemID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "item_id", itemID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.DeleteTemplateItemResponse{})
}

func (h *Handler) HandleUseTemplate(msg *nats.Msg) {
	correlationID := msg.Header.Get("X-Correlation-ID")
	ctx := applogger.WithCorrelationID(context.Background(), correlationID)

	var req todov1.UseTemplateRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.WarnContext(ctx, "failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	templateID, err := uuid.Parse(req.TemplateId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid template_id", "subject", msg.Subject, "template_id", req.TemplateId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid template ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user_id", "subject", msg.Subject, "user_id", req.UserId)
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	list, err := h.service.UseTemplate(ctx, templateID, userID, req.Title)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.WarnContext(ctx, "validation failed", "error", err, "template_id", templateID, "user_id", userID)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.ErrorContext(ctx, "service error", "error", err, "template_id", templateID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.UseTemplateResponse{
		List: &todov1.List{
			Id:        list.ID.String(),
			UserId:    list.UserID.String(),
			Title:     list.Title,
			CreatedAt: timestamppb.New(list.CreatedAt),
			UpdatedAt: timestamppb.New(list.UpdatedAt),
		},
	})
}
