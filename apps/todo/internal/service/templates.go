package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/terjelafton/platform/apps/todo/internal/db"
)

func (s *Service) CreateTemplate(
	ctx context.Context,
	userID uuid.UUID,
	title string,
) (*db.TodoTemplate, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrTitleRequired
	}

	if len(title) > 100 {
		return nil, ErrTitleTooLong
	}

	template, err := s.queries.CreateTemplate(ctx, db.CreateTemplateParams{
		UserID: userID,
		Title:  title,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "template created", "template_id", template.ID, "user_id", userID)

	return &template, nil
}

func (s *Service) GetTemplatesByUser(ctx context.Context, userID uuid.UUID) ([]db.GetTemplatesByUserRow, error) {
	templates, err := s.queries.GetTemplatesByUser(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "templates retrieved", "user_id", userID, "count", len(templates))

	return templates, nil
}

func (s *Service) UpdateTemplateTitle(
	ctx context.Context,
	templateID, userID uuid.UUID,
	title string,
) (*db.TodoTemplate, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrTitleRequired
	}

	if len(title) > 100 {
		return nil, ErrTitleTooLong
	}

	template, err := s.queries.UpdateTemplateTitle(ctx, db.UpdateTemplateTitleParams{
		ID:     templateID,
		UserID: userID,
		Title:  title,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "template_id", templateID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "template title updated", "template_id", templateID, "user_id", userID)

	return &template, nil
}

func (s *Service) DeleteTemplate(ctx context.Context, templateID, userID uuid.UUID) error {
	_, err := s.queries.DeleteTemplate(ctx, db.DeleteTemplateParams{
		ID:     templateID,
		UserID: userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "template_id", templateID, "user_id", userID)
		return translateDBError(err)
	}

	s.logger.InfoContext(ctx, "template deleted", "template_id", templateID, "user_id", userID)

	return nil
}

func (s *Service) CreateTemplateItem(
	ctx context.Context,
	templateID, userID uuid.UUID,
	title string,
) (*db.TodoTemplateItem, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrTitleRequired
	}

	if len(title) > 500 {
		return nil, ErrItemTitleTooLong
	}

	item, err := s.queries.CreateTemplateItem(ctx, db.CreateTemplateItemParams{
		TemplateID: templateID,
		Title:      title,
		UserID:     userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "template_id", templateID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "template item created", "item_id", item.ID, "template_id", templateID, "user_id", userID)

	return &item, nil
}

func (s *Service) GetTemplateItems(
	ctx context.Context,
	templateID, userID uuid.UUID,
) ([]db.TodoTemplateItem, error) {
	items, err := s.queries.GetTemplateItems(ctx, db.GetTemplateItemsParams{
		TemplateID: templateID,
		UserID:     userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "template_id", templateID, "user_id", userID)
		return nil, translateDBError(err)
	}

	s.logger.InfoContext(ctx, "template items retrieved", "template_id", templateID, "user_id", userID, "count", len(items))

	return items, nil
}

func (s *Service) DeleteTemplateItem(ctx context.Context, itemID, userID uuid.UUID) error {
	_, err := s.queries.DeleteTemplateItem(ctx, db.DeleteTemplateItemParams{
		ID:     itemID,
		UserID: userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "item_id", itemID, "user_id", userID)
		return translateDBError(err)
	}

	s.logger.InfoContext(ctx, "template item deleted", "item_id", itemID, "user_id", userID)

	return nil
}

func (s *Service) UseTemplate(
	ctx context.Context,
	templateID, userID uuid.UUID,
	title string,
) (*db.TodoList, error) {
	// Get template (verifies ownership via WHERE user_id = $2)
	template, err := s.queries.GetTemplate(ctx, db.GetTemplateParams{
		ID:     templateID,
		UserID: userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "template_id", templateID, "user_id", userID)
		return nil, translateDBError(err)
	}

	// Fall back to template title if no title provided
	title = strings.TrimSpace(title)
	if title == "" {
		title = template.Title
	}

	if len(title) > 100 {
		return nil, ErrTitleTooLong
	}

	// Create the list
	list, err := s.queries.CreateList(ctx, db.CreateListParams{
		UserID: userID,
		Title:  title,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "template_id", templateID, "user_id", userID)
		return nil, translateDBError(err)
	}

	// Get template items
	templateItems, err := s.queries.GetTemplateItems(ctx, db.GetTemplateItemsParams{
		TemplateID: templateID,
		UserID:     userID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "database error", "error", err, "template_id", templateID, "user_id", userID)
		return nil, translateDBError(err)
	}

	// Create each template item as a list item
	for _, ti := range templateItems {
		_, err := s.queries.CreateItem(ctx, db.CreateItemParams{
			ListID: list.ID,
			Title:  ti.Title,
			UserID: userID,
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "database error", "error", err, "template_id", templateID, "list_id", list.ID, "user_id", userID)
			return nil, translateDBError(err)
		}
	}

	s.logger.InfoContext(ctx, "template used", "template_id", templateID, "list_id", list.ID, "user_id", userID, "items_count", len(templateItems))

	return &list, nil
}
