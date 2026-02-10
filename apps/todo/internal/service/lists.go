package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/terjelafton/platform/apps/todo/internal/db"
	todov1 "github.com/terjelafton/platform/libs/proto-stubs/todo/v1"
)

func (s *Service) CreateList(
	ctx context.Context,
	userID uuid.UUID,
	title string,
) (*db.TodoList, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrTitleRequired
	}

	if len(title) > 100 {
		return nil, ErrTitleTooLong
	}

	list, err := s.queries.CreateList(ctx, db.CreateListParams{
		UserID: userID,
		Title:  title,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "list created", "list_id", list.ID, "user_id", userID)

	s.publishListCreated(ctx, list)

	return &list, nil
}

func (s *Service) publishListCreated(ctx context.Context, list db.TodoList) {
	event := &todov1.ListCreatedEvent{
		List: &todov1.List{
			Id:        list.ID.String(),
			UserId:    list.UserID.String(),
			Title:     list.Title,
			CreatedAt: timestamppb.New(list.CreatedAt),
			UpdatedAt: timestamppb.New(list.UpdatedAt),
		},
		OccurredAt: timestamppb.Now(),
	}

	data, err := proto.Marshal(event)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to marshal event", "error", err, "list_id", list.ID)
		return
	}

	if err := s.nc.Publish("todo.list.created", data); err != nil {
		s.logger.WarnContext(ctx, "failed to publish event", "error", err, "list_id", list.ID)
	} else {
		s.logger.InfoContext(ctx, "event published", "subject", "todo.list.created", "list_id", list.ID)
	}
}

func (s *Service) GetListsByUser(ctx context.Context, userID uuid.UUID) ([]db.GetListsByUserRow, error) {
	lists, err := s.queries.GetListsByUser(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "lists retrieved", "user_id", userID, "count", len(lists))

	return lists, nil
}

func (s *Service) DeleteList(ctx context.Context, listID, userID uuid.UUID) error {
	_, err := s.queries.DeleteList(ctx, db.DeleteListParams{
		ID:     listID,
		UserID: userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "list_id", listID, "user_id", userID)
		return translateDBError(err)
	}

	s.logger.InfoContext(ctx, "list deleted", "list_id", listID, "user_id", userID)

	return nil
}

func (s *Service) UpdateListTitle(
	ctx context.Context,
	listID, userID uuid.UUID,
	title string,
) (*db.TodoList, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrTitleRequired
	}

	if len(title) > 100 {
		return nil, ErrTitleTooLong
	}

	list, err := s.queries.UpdateListTitle(ctx, db.UpdateListTitleParams{
		ID:     listID,
		UserID: userID,
		Title:  title,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "list_id", listID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "list title updated", "list_id", listID, "user_id", userID)

	return &list, nil
}
