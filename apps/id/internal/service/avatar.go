package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/terjelafton/platform/apps/id/internal/db"
)

const maxAvatarSize = 1 << 20 // 1MB

var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

func (s *Service) GetAvatar(ctx context.Context, userID uuid.UUID) ([]byte, string, error) {
	row, err := s.queries.GetUserAvatar(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get avatar", "error", err, "user_id", userID)
		return nil, "", fmt.Errorf("internal error")
	}
	if row.Avatar == nil {
		return nil, "", nil
	}
	contentType := ""
	if row.AvatarContentType.Valid {
		contentType = row.AvatarContentType.String
	}
	return row.Avatar, contentType, nil
}

func (s *Service) UpdateAvatar(ctx context.Context, userID uuid.UUID, data []byte, contentType string) error {
	if len(data) == 0 {
		return &ValidationError{Field: "avatar", Message: "Avatar data is required"}
	}
	if len(data) > maxAvatarSize {
		return &ValidationError{Field: "avatar", Message: "Avatar must be 1MB or less"}
	}
	if !allowedContentTypes[contentType] {
		return &ValidationError{Field: "avatar", Message: "Avatar must be JPEG or PNG"}
	}

	err := s.queries.UpdateUserAvatar(ctx, db.UpdateUserAvatarParams{
		ID:                userID,
		Avatar:            data,
		AvatarContentType: sql.NullString{String: contentType, Valid: true},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update avatar", "error", err, "user_id", userID)
		return fmt.Errorf("internal error")
	}

	s.logger.InfoContext(ctx, "avatar updated", "user_id", userID)
	return nil
}
