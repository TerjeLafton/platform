package nats

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/terjelafton/platform/apps/id/internal/service"
	idv1 "github.com/terjelafton/platform/libs/proto-stubs/id/v1"
	"google.golang.org/protobuf/proto"
)

func (h *Handler) HandleRegister(msg *nats.Msg) {
	var req idv1.RegisterRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Warn("failed to unmarshal register request", "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	user, token, err := h.service.Register(context.Background(), req.Email, req.Password, req.Name)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.Warn("validation failed", "field", valErr.Field, "error", valErr.Message)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Error())
			return
		}
		h.logger.Error("registration failed", "error", err)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	resp := &idv1.RegisterResponse{
		User: &idv1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Name:      user.Name,
		},
		Token: token,
	}

	h.respondSuccess(msg, resp)
}

func (h *Handler) HandleLogin(msg *nats.Msg) {
	var req idv1.LoginRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Warn("failed to unmarshal login request", "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	user, token, err := h.service.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		if valErr, ok := err.(*service.ValidationError); ok {
			h.logger.Warn("validation failed", "field", valErr.Field, "error", valErr.Message)
			h.respondError(msg, "VALIDATION_ERROR", valErr.Error())
			return
		}
		if authErr, ok := err.(*service.AuthError); ok {
			h.logger.Warn("authentication failed", "error", authErr.Message)
			h.respondError(msg, "AUTH_ERROR", authErr.Error())
			return
		}
		h.logger.Error("login failed", "error", err)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	resp := &idv1.LoginResponse{
		User: &idv1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Name:      user.Name,
		},
		Token: token,
	}

	h.respondSuccess(msg, resp)
}

func (h *Handler) HandleValidateToken(msg *nats.Msg) {
	var req idv1.ValidateTokenRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Warn("failed to unmarshal validate request", "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	userID, err := h.service.ValidateToken(req.Token)
	if err != nil {
		h.logger.Warn("token validation failed", "error", err)
		h.respondError(msg, "AUTH_ERROR", "Invalid or expired token")
		return
	}

	resp := &idv1.ValidateTokenResponse{
		UserId: userID.String(),
	}

	h.logger.Info("token validated", "user_id", userID)
	h.respondSuccess(msg, resp)
}

func (h *Handler) HandleGetUser(msg *nats.Msg) {
	var req idv1.GetUserRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.logger.Warn("failed to unmarshal get user request", "error", err)
		h.respondError(msg, "INVALID_REQUEST", "Invalid request format")
		return
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		h.logger.Warn("invalid user ID", "user_id", req.UserId)
		h.respondError(msg, "VALIDATION_ERROR", "Invalid user ID")
		return
	}

	user, err := h.service.GetUser(context.Background(), userID)
	if err != nil {
		h.logger.Error("failed to get user", "error", err, "user_id", req.UserId)
		h.respondError(msg, "INTERNAL_ERROR", "Internal server error")
		return
	}

	resp := &idv1.GetUserResponse{
		User: &idv1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}
	h.respondSuccess(msg, resp)
}
