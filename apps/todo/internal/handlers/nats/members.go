package nats

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/terjelafton/platform/apps/todo/internal/service"
	todov1 "github.com/terjelafton/platform/libs/proto-stubs/todo/v1"
)

func (h *Handler) HandleAddListMember(msg *nats.Msg) {
	var req todov1.AddListMemberRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Warn("failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	listID, err := uuid.Parse(req.ListId)
	if err != nil {
		h.respondError(msg, "INVALID_REQUEST", "Invalid list ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	member, err := h.service.AddListMember(context.Background(), listID, userID, req.MemberEmail)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.Error("service error", "error", err, "list_id", listID, "user_id", userID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.AddListMemberResponse{
		Member: &todov1.ListMember{
			UserId:  member.UserID.String(),
			Email:   member.Email,
			Name:    member.Name,
			AddedAt: timestamppb.New(member.CreatedAt),
		},
	})
}

func (h *Handler) HandleRemoveListMember(msg *nats.Msg) {
	var req todov1.RemoveListMemberRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Warn("failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	listID, err := uuid.Parse(req.ListId)
	if err != nil {
		h.respondError(msg, "INVALID_REQUEST", "Invalid list ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	memberUserID, err := uuid.Parse(req.MemberUserId)
	if err != nil {
		h.respondError(msg, "INVALID_REQUEST", "Invalid member user ID format")
		return
	}

	if err := h.service.RemoveListMember(context.Background(), listID, userID, memberUserID); err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.Error("service error", "error", err, "list_id", listID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	h.respondSuccess(msg, &todov1.RemoveListMemberResponse{})
}

func (h *Handler) HandleGetListMembers(msg *nats.Msg) {
	var req todov1.GetListMembersRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Warn("failed to unmarshal request", "subject", msg.Subject, "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	listID, err := uuid.Parse(req.ListId)
	if err != nil {
		h.respondError(msg, "INVALID_REQUEST", "Invalid list ID format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.respondError(msg, "INVALID_REQUEST", "Invalid user ID format")
		return
	}

	members, err := h.service.GetListMembers(context.Background(), listID, userID)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.respondError(msg, "VALIDATION_ERROR", valErr.Message)
		} else {
			h.logger.Error("service error", "error", err, "list_id", listID)
			h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	pbMembers := make([]*todov1.ListMember, len(members))
	for i, m := range members {
		pbMembers[i] = &todov1.ListMember{
			UserId:  m.UserID.String(),
			Email:   m.Email,
			Name:    m.Name,
			AddedAt: timestamppb.New(m.CreatedAt),
		}
	}

	h.respondSuccess(msg, &todov1.GetListMembersResponse{
		Members: pbMembers,
	})
}
