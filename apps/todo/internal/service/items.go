package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/terjelafton/platform/apps/todo/internal/db"
	todov1 "github.com/terjelafton/platform/libs/proto-stubs/todo/v1"
)

func (s *Service) CreateItem(
	ctx context.Context,
	listID uuid.UUID,
	userID uuid.UUID,
	title string,
	correlationID string,
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
		s.logger.ErrorContext(ctx, "database error", "error", err, "list_id", listID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "item created", "item_id", item.ID, "list_id", listID, "user_id", userID)

	s.publishItemEvent(ctx, item.ListID.String(), "item.created", correlationID, &todov1.Item{
		Id:        item.ID.String(),
		ListId:    item.ListID.String(),
		Title:     item.Title,
		Completed: item.Completed,
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	})

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
		s.logger.ErrorContext(ctx, "database error", "error", err, "list_id", listID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "items retrieved", "list_id", listID, "user_id", userID, "count", len(items))

	return items, nil
}

func (s *Service) ToggleItemCompleted(
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
	correlationID string,
) (*db.TodoItem, error) {
	item, err := s.queries.ToggleItemCompleted(ctx, db.ToggleItemCompletedParams{
		ID:     itemID,
		UserID: userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "item_id", itemID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "item toggled", "item_id", itemID, "user_id", userID, "completed", item.Completed)

	s.publishItemEvent(ctx, item.ListID.String(), "item.toggled", correlationID, &todov1.Item{
		Id:        item.ID.String(),
		ListId:    item.ListID.String(),
		Title:     item.Title,
		Completed: item.Completed,
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	})

	return &item, nil
}

func (s *Service) DeleteItem(ctx context.Context, itemID, userID, listID uuid.UUID, correlationID string) error {
	_, err := s.queries.DeleteItem(ctx, db.DeleteItemParams{
		ID:     itemID,
		UserID: userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "item_id", itemID, "user_id", userID)
		return translateDBError(err)
	}

	s.logger.InfoContext(ctx, "item deleted", "item_id", itemID, "user_id", userID)

	s.publishItemEvent(ctx, listID.String(), "item.deleted", correlationID, &todov1.ItemDeletedEvent{
		ListId: listID.String(),
		ItemId: itemID.String(),
	})

	return nil
}

func (s *Service) publishItemEvent(ctx context.Context, listID, eventType, correlationID string, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to marshal event", "error", err, "list_id", listID, "event_type", eventType)
		return
	}

	natsMsg := &nats.Msg{
		Subject: fmt.Sprintf("todo.events.%s.%s", listID, eventType),
		Data:    data,
	}
	if correlationID != "" {
		natsMsg.Header = nats.Header{}
		natsMsg.Header.Set("X-Correlation-ID", correlationID)
	}

	if err := s.nc.PublishMsg(natsMsg); err != nil {
		s.logger.WarnContext(ctx, "failed to publish event", "error", err, "list_id", listID, "event_type", eventType)
	} else {
		s.logger.InfoContext(ctx, "event published", "subject", natsMsg.Subject, "list_id", listID)
	}
}
