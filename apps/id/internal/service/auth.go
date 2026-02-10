package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/terjelafton/platform/apps/id/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(ctx context.Context, email, password, name string) (*db.IDUser, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)

	if email == "" {
		return nil, "", &ValidationError{Field: "email", Message: "Email is required"}
	}
	if !strings.Contains(email, "@") {
		return nil, "", &ValidationError{Field: "email", Message: "Invalid email format"}
	}
	if len(password) < 8 {
		return nil, "", &ValidationError{Field: "password", Message: "Password must be at least 8 characters"}
	}
	if name == "" {
		return nil, "", &ValidationError{Field: "name", Message: "Name is required"}
	}
	if len(name) > 100 {
		return nil, "", &ValidationError{Field: "name", Message: "Name must be 100 characters or less"}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to hash password", "error", err)
		return nil, "", fmt.Errorf("internal error")
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, "", &ValidationError{Field: "email", Message: "Email already registered"}
		}
		s.logger.ErrorContext(ctx, "failed to create user", "error", err)
		return nil, "", fmt.Errorf("internal error")
	}

	token, err := s.CreateToken(user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create token", "error", err, "user_id", user.ID)
		return nil, "", fmt.Errorf("internal error")
	}

	s.logger.InfoContext(ctx, "user registered", "user_id", user.ID, "email", email)
	return &user, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*db.IDUser, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, "", &ValidationError{Field: "email", Message: "Email is required"}
	}
	if password == "" {
		return nil, "", &ValidationError{Field: "password", Message: "Password is required"}
	}

	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.WarnContext(ctx, "login attempt for non-existent email", "email", email)
			return nil, "", &AuthError{Message: "Invalid email or password"}
		}
		s.logger.ErrorContext(ctx, "failed to get user", "error", err, "email", email)
		return nil, "", fmt.Errorf("internal error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.WarnContext(ctx, "login attempt with wrong password", "email", email)
		return nil, "", &AuthError{Message: "Invalid email or password"}
	}

	token, err := s.CreateToken(user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create token", "error", err, "user_id", user.ID)
		return nil, "", fmt.Errorf("internal error")
	}

	s.logger.InfoContext(ctx, "user logged in", "user_id", user.ID, "email", email)
	return &user, token, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*db.IDUser, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, &ValidationError{Field: "email", Message: "Email is required"}
	}

	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ValidationError{Field: "email", Message: "No account found with that email"}
		}
		s.logger.ErrorContext(ctx, "failed to get user by email", "error", err, "email", email)
		return nil, fmt.Errorf("internal error")
	}

	return &user, nil
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*db.IDUser, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user", "error", err, "user_id", userID)
		return nil, fmt.Errorf("internal error")
	}
	return &user, nil
}
