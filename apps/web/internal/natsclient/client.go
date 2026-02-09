package natsclient

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	commonv1 "github.com/terjelafton/platform/libs/proto-stubs/common/v1"
	idv1 "github.com/terjelafton/platform/libs/proto-stubs/id/v1"
	"google.golang.org/protobuf/proto"
)

const requestTimeout = 5 * time.Second

type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func Login(nc *nats.Conn, email, password string) (string, error) {
	req := &idv1.LoginRequest{
		Email:    email,
		Password: password,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal login request: %w", err)
	}

	msg, err := nc.Request("id.auth.login", data, requestTimeout)
	if err != nil {
		return "", fmt.Errorf("nats request: %w", err)
	}

	// Try as success response — check that token is actually populated
	var resp idv1.LoginResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.Token != "" {
		return resp.Token, nil
	}

	// Try as error response
	if svcErr := parseError(msg.Data); svcErr != nil {
		return "", svcErr
	}

	return "", fmt.Errorf("unexpected response from id service")
}

func Register(nc *nats.Conn, email, password, name string) (string, error) {
	req := &idv1.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal register request: %w", err)
	}

	msg, err := nc.Request("id.auth.register", data, requestTimeout)
	if err != nil {
		return "", fmt.Errorf("nats request: %w", err)
	}

	// Try as success response
	var resp idv1.RegisterResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.Token != "" {
		return resp.Token, nil
	}

	// Try as error response
	if svcErr := parseError(msg.Data); svcErr != nil {
		return "", svcErr
	}

	return "", fmt.Errorf("unexpected response from id service")
}

func ValidateToken(nc *nats.Conn, token string) (string, error) {
	req := &idv1.ValidateTokenRequest{
		Token: token,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal validate request: %w", err)
	}

	msg, err := nc.Request("id.auth.validate", data, requestTimeout)
	if err != nil {
		return "", fmt.Errorf("nats request: %w", err)
	}

	// Try as success response — check that user_id is populated
	var resp idv1.ValidateTokenResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.UserId != "" {
		return resp.UserId, nil
	}

	// Try as error response
	if svcErr := parseError(msg.Data); svcErr != nil {
		return "", svcErr
	}

	return "", fmt.Errorf("unexpected response from id service")
}

func GetUser(nc *nats.Conn, userID string) (*idv1.User, error) {
	req := &idv1.GetUserRequest{
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get user request: %w", err)
	}

	msg, err := nc.Request("id.user.get", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp idv1.GetUserResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.User != nil {
		return resp.User, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from id service")
}

func GetAvatar(nc *nats.Conn, userID string) ([]byte, string, error) {
	req := &idv1.GetAvatarRequest{
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, "", fmt.Errorf("marshal get avatar request: %w", err)
	}

	msg, err := nc.Request("id.user.avatar", data, requestTimeout)
	if err != nil {
		return nil, "", fmt.Errorf("nats request: %w", err)
	}

	var resp idv1.GetAvatarResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && len(resp.Avatar) > 0 {
		return resp.Avatar, resp.ContentType, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, "", svcErr
	}

	// No avatar set — return nil (not an error)
	return nil, "", nil
}

func UpdateAvatar(nc *nats.Conn, userID string, avatar []byte, contentType string) error {
	req := &idv1.UpdateAvatarRequest{
		UserId:      userID,
		Avatar:      avatar,
		ContentType: contentType,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal update avatar request: %w", err)
	}

	msg, err := nc.Request("id.user.avatar.update", data, requestTimeout)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func GetUserByEmail(nc *nats.Conn, email string) (*idv1.User, error) {
	req := &idv1.GetUserByEmailRequest{
		Email: email,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get user by email request: %w", err)
	}

	msg, err := nc.Request("id.user.get_by_email", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp idv1.GetUserByEmailResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.User != nil {
		return resp.User, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from id service")
}

func parseError(data []byte) *ServiceError {
	var errResp commonv1.ErrorResponse
	if err := proto.Unmarshal(data, &errResp); err == nil && errResp.Code != "" {
		return &ServiceError{
			Code:    errResp.Code,
			Message: errResp.Message,
		}
	}
	return nil
}
