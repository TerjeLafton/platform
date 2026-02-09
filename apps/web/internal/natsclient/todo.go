package natsclient

import (
	"fmt"

	"github.com/nats-io/nats.go"
	todov1 "github.com/terjelafton/platform/libs/proto-stubs/todo/v1"
	"google.golang.org/protobuf/proto"
)

func CreateList(nc *nats.Conn, userID, title string) (*todov1.List, error) {
	req := &todov1.CreateListRequest{
		UserId: userID,
		Title:  title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal create list request: %w", err)
	}

	msg, err := nc.Request("todo.list.create", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.CreateListResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.List != nil {
		return resp.List, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func GetListsByUser(nc *nats.Conn, userID string) ([]*todov1.List, error) {
	req := &todov1.GetListsByUserRequest{
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get lists request: %w", err)
	}

	msg, err := nc.Request("todo.list.get_by_user", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.GetListsByUserResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil {
		return resp.Lists, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func UpdateListTitle(nc *nats.Conn, id, userID, title string) (*todov1.List, error) {
	req := &todov1.UpdateListTitleRequest{
		Id:     id,
		UserId: userID,
		Title:  title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal update list title request: %w", err)
	}

	msg, err := nc.Request("todo.list.update_title", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.UpdateListTitleResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.List != nil {
		return resp.List, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func DeleteList(nc *nats.Conn, id, userID string) error {
	req := &todov1.DeleteListRequest{
		Id:     id,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal delete list request: %w", err)
	}

	msg, err := nc.Request("todo.list.delete", data, requestTimeout)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	// DeleteListResponse is empty on success — only check for errors
	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func CreateItem(nc *nats.Conn, listID, userID, title string) (*todov1.Item, error) {
	req := &todov1.CreateItemRequest{
		ListId: listID,
		UserId: userID,
		Title:  title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal create item request: %w", err)
	}

	msg, err := nc.Request("todo.item.create", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.CreateItemResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.Item != nil {
		return resp.Item, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func GetAllItemsFromList(nc *nats.Conn, listID, userID string) ([]*todov1.Item, error) {
	req := &todov1.GetAllItemsFromListRequest{
		ListId: listID,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get items request: %w", err)
	}

	msg, err := nc.Request("todo.item.get_all", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.GetAllItemsFromListResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil {
		return resp.Items, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func ToggleItemCompleted(nc *nats.Conn, id, userID string) (*todov1.Item, error) {
	req := &todov1.ToggleItemCompletedRequest{
		Id:     id,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal toggle item request: %w", err)
	}

	msg, err := nc.Request("todo.item.toggle_completed", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.ToggleItemCompletedResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.Item != nil {
		return resp.Item, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func DeleteItem(nc *nats.Conn, id, userID string) error {
	req := &todov1.DeleteItemRequest{
		Id:     id,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal delete item request: %w", err)
	}

	msg, err := nc.Request("todo.item.delete", data, requestTimeout)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func AddListMember(nc *nats.Conn, listID, userID, memberEmail string) (*todov1.ListMember, error) {
	req := &todov1.AddListMemberRequest{
		ListId:      listID,
		UserId:      userID,
		MemberEmail: memberEmail,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal add list member request: %w", err)
	}

	msg, err := nc.Request("todo.list.add_member", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.AddListMemberResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.Member != nil {
		return resp.Member, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func RemoveListMember(nc *nats.Conn, listID, userID, memberUserID string) error {
	req := &todov1.RemoveListMemberRequest{
		ListId:       listID,
		UserId:       userID,
		MemberUserId: memberUserID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal remove list member request: %w", err)
	}

	msg, err := nc.Request("todo.list.remove_member", data, requestTimeout)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func GetListMembers(nc *nats.Conn, listID, userID string) ([]*todov1.ListMember, error) {
	req := &todov1.GetListMembersRequest{
		ListId: listID,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get list members request: %w", err)
	}

	msg, err := nc.Request("todo.list.get_members", data, requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.GetListMembersResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil {
		return resp.Members, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}
