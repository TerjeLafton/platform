package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	commonv1 "github.com/terjelafton/platform/libs/proto-stubs/common/v1"
	idv1 "github.com/terjelafton/platform/libs/proto-stubs/id/v1"
	"google.golang.org/protobuf/proto"

	"github.com/terjelafton/platform/apps/todo/internal/db"
)

func (s *Service) AddListMember(
	ctx context.Context,
	listID, ownerUserID uuid.UUID,
	memberEmail string,
) (*db.TodoListMember, error) {
	memberEmail = strings.TrimSpace(strings.ToLower(memberEmail))
	if memberEmail == "" {
		return nil, ErrEmailRequired
	}

	// Verify requesting user is the list owner
	ownerID, err := s.queries.GetListOwner(ctx, listID)
	if err != nil {
		s.logger.Error("database error", "error", err, "list_id", listID)
		return nil, translateDBError(err)
	}
	if ownerID != ownerUserID {
		return nil, ErrNotListOwner
	}

	// Look up the member by email via id service
	user, err := s.getUserByEmail(memberEmail)
	if err != nil {
		return nil, err
	}

	memberUserID, err := uuid.Parse(user.Id)
	if err != nil {
		s.logger.Error("invalid user ID from id service", "error", err, "user_id", user.Id)
		return nil, fmt.Errorf("internal error")
	}

	// Cannot share with yourself
	if memberUserID == ownerUserID {
		return nil, ErrCannotShareWithSelf
	}

	member, err := s.queries.AddListMember(ctx, db.AddListMemberParams{
		ListID: listID,
		UserID: memberUserID,
		Name:   user.Name,
		Email:  user.Email,
	})
	if err != nil {
		s.logger.Error("database error", "error", err, "list_id", listID, "member_user_id", memberUserID)
		return nil, translateDBError(err)
	}

	s.logger.Info("member added to list", "list_id", listID, "member_user_id", memberUserID, "owner_user_id", ownerUserID)

	return &member, nil
}

func (s *Service) RemoveListMember(
	ctx context.Context,
	listID, requestingUserID, memberUserID uuid.UUID,
) error {
	ownerID, err := s.queries.GetListOwner(ctx, listID)
	if err != nil {
		s.logger.Error("database error", "error", err, "list_id", listID)
		return translateDBError(err)
	}

	if requestingUserID == memberUserID {
		// Member leaving — verify they're a member
		isMember, err := s.queries.IsListMember(ctx, db.IsListMemberParams{
			ListID: listID,
			UserID: memberUserID,
		})
		if err != nil {
			s.logger.Error("database error", "error", err, "list_id", listID)
			return translateDBError(err)
		}
		if !isMember {
			return ErrNotListMember
		}
	} else {
		// Owner removing someone
		if ownerID != requestingUserID {
			return ErrNotListOwner
		}
	}

	if err := s.queries.RemoveListMember(ctx, db.RemoveListMemberParams{
		ListID: listID,
		UserID: memberUserID,
	}); err != nil {
		s.logger.Error("database error", "error", err, "list_id", listID, "member_user_id", memberUserID)
		return translateDBError(err)
	}

	s.logger.Info("member removed from list", "list_id", listID, "member_user_id", memberUserID, "requested_by", requestingUserID)

	return nil
}

func (s *Service) GetListMembers(
	ctx context.Context,
	listID, userID uuid.UUID,
) ([]db.TodoListMember, error) {
	// Verify user is owner or member
	ownerID, err := s.queries.GetListOwner(ctx, listID)
	if err != nil {
		s.logger.Error("database error", "error", err, "list_id", listID)
		return nil, translateDBError(err)
	}

	if ownerID != userID {
		isMember, err := s.queries.IsListMember(ctx, db.IsListMemberParams{
			ListID: listID,
			UserID: userID,
		})
		if err != nil {
			s.logger.Error("database error", "error", err, "list_id", listID)
			return nil, translateDBError(err)
		}
		if !isMember {
			return nil, ErrNotListMember
		}
	}

	members, err := s.queries.GetListMembers(ctx, listID)
	if err != nil {
		s.logger.Error("database error", "error", err, "list_id", listID)
		return nil, translateDBError(err)
	}

	s.logger.Info("members retrieved", "list_id", listID, "count", len(members))

	return members, nil
}

// getUserByEmail calls the id service via NATS to look up a user by email.
func (s *Service) getUserByEmail(email string) (*idv1.User, error) {
	req := &idv1.GetUserByEmailRequest{Email: email}
	data, err := proto.Marshal(req)
	if err != nil {
		s.logger.Error("failed to marshal request", "error", err)
		return nil, fmt.Errorf("internal error")
	}

	msg, err := s.nc.Request("id.user.get_by_email", data, 5*time.Second)
	if err != nil {
		s.logger.Error("NATS request failed", "error", err, "subject", "id.user.get_by_email")
		return nil, fmt.Errorf("internal error")
	}

	var resp idv1.GetUserByEmailResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.User != nil {
		return resp.User, nil
	}

	// Check for error response
	var errResp commonv1.ErrorResponse
	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		if errResp.Code == "VALIDATION_ERROR" {
			return nil, &ValidationError{Field: "email", Message: errResp.Message}
		}
		return nil, fmt.Errorf("id service error: %s", errResp.Message)
	}

	return nil, fmt.Errorf("unexpected response from id service")
}
