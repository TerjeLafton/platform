package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/terjelafton/platform/apps/todo/internal/db"
)

func (s *Service) CreateItem(
	ctx context.Context,
	listID uuid.UUID,
	userID uuid.UUID,
	title string,
) (*db.TodoItem, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrTitleRequired
	}

	if len(title) > 500 {
		return nil, ErrItemTitleTooLong
	}

	item, err := s.queries.CreateItem(ctx, db.CreateItemParams{
		ListID: listID,
		Title:  title,
		UserID: userID,
	})
	if err != nil {
		s.logger.Error("database error", "error", err, "list_id", listID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.Info("item created", "item_id", item.ID, "list_id", listID, "user_id", userID)

	return &item, nil
}

func (s *Service) GetAllItemsFromList(
	ctx context.Context,
	listID uuid.UUID,
	userID uuid.UUID,
) ([]db.TodoItem, error) {
	items, err := s.queries.GetAllItemsFromList(ctx, db.GetAllItemsFromListParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		s.logger.Error("database error", "error", err, "list_id", listID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.Info("items retrieved", "list_id", listID, "user_id", userID, "count", len(items))

	return items, nil
}

func (s *Service) ToggleItemCompleted(
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
) (*db.TodoItem, error) {
	item, err := s.queries.ToggleItemCompleted(ctx, db.ToggleItemCompletedParams{
		ID:     itemID,
		UserID: userID,
	})
	if err != nil {
		s.logger.Error("database error", "error", err, "item_id", itemID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.Info("item toggled", "item_id", itemID, "user_id", userID, "completed", item.Completed)

	return &item, nil
}

func (s *Service) DeleteItem(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) error {
	_, err := s.queries.DeleteItem(ctx, db.DeleteItemParams{
		ID:     itemID,
		UserID: userID,
	})
	if err != nil {
		s.logger.Error("database error", "error", err, "item_id", itemID, "user_id", userID)
		return translateDBError(err)
	}

	s.logger.Info("item deleted", "item_id", itemID, "user_id", userID)

	return nil
}
