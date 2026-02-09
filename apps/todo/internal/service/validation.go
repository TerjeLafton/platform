package service

import "strings"

type ValidationError struct {
	Field   string
	Message string
}

var (
	ErrTitleRequired = &ValidationError{
		Field:   "title",
		Message: "Title is required",
	}
	ErrTitleTooLong = &ValidationError{
		Field:   "title",
		Message: "Title must be less than 100 characters",
	}
	ErrItemTitleTooLong = &ValidationError{
		Field:   "title",
		Message: "Item title must be less than 500 characters",
	}
	ErrEmailRequired = &ValidationError{
		Field:   "email",
		Message: "Email is required",
	}
	ErrCannotShareWithSelf = &ValidationError{
		Field:   "email",
		Message: "You can't share a list with yourself",
	}
	ErrNotListOwner = &ValidationError{
		Field:   "user_id",
		Message: "Only the list owner can do this",
	}
	ErrNotTemplateOwner = &ValidationError{
		Field:   "user_id",
		Message: "Only the template owner can do this",
	}
	ErrNotListMember = &ValidationError{
		Field:   "user_id",
		Message: "You are not a member of this list",
	}
	ErrAlreadyMember = &ValidationError{
		Field:   "email",
		Message: "This user is already a member of this list",
	}
)

func (e *ValidationError) Error() string {
	return e.Message
}

func translateDBError(err error) error {
	errStr := err.Error()

	// PostgreSQL unique constraint violation (templates)
	if strings.Contains(errStr, "templates_user_id_title_key") {
		return &ValidationError{
			Field:   "title",
			Message: "You already have a template with this title",
		}
	}

	// PostgreSQL unique constraint violation (lists)
	if strings.Contains(errStr, "duplicate key") && strings.Contains(errStr, "user_id") {
		return &ValidationError{
			Field:   "title",
			Message: "You already have a list with this title",
		}
	}

	// PostgreSQL CHECK constraint violations
	if strings.Contains(errStr, "templates_title_check") {
		return &ValidationError{
			Field:   "title",
			Message: "Title is invalid (must be 1-100 characters)",
		}
	}
	if strings.Contains(errStr, "template_items_title_check") {
		return &ValidationError{
			Field:   "title",
			Message: "Title is invalid (must be 1-500 characters)",
		}
	}
	if strings.Contains(errStr, "lists_title_check") {
		return &ValidationError{
			Field:   "title",
			Message: "Title is invalid (must be 1-100 characters)",
		}
	}
	if strings.Contains(errStr, "items_title_check") {
		return &ValidationError{
			Field:   "title",
			Message: "Title is invalid (must be 1-500 characters)",
		}
	}

	// PostgreSQL foreign key violation
	if strings.Contains(errStr, "violates foreign key constraint") {
		return &ValidationError{
			Field:   "list_id",
			Message: "List not found",
		}
	}

	// Duplicate list member
	if strings.Contains(errStr, "list_members_pkey") {
		return &ValidationError{
			Field:   "email",
			Message: "This user is already a member of this list",
		}
	}

	// No rows returned (for operations that should affect exactly one row)
	if strings.Contains(errStr, "no rows") {
		return &ValidationError{
			Field:   "id",
			Message: "Resource not found or you don't have permission",
		}
	}

	return err
}
