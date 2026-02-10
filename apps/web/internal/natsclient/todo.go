package natsclient

import (
	"fmt"

	"github.com/nats-io/nats.go"
	todov1 "github.com/terjelafton/platform/libs/proto-stubs/todo/v1"
	"google.golang.org/protobuf/proto"
)

func CreateList(nc *nats.Conn, userID, title, correlationID string) (*todov1.List, error) {
	req := &todov1.CreateListRequest{
		UserId: userID,
		Title:  title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal create list request: %w", err)
	}

	msg, err := request(nc, "todo.list.create", data, correlationID)
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

func GetListsByUser(nc *nats.Conn, userID, correlationID string) ([]*todov1.List, error) {
	req := &todov1.GetListsByUserRequest{
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get lists request: %w", err)
	}

	msg, err := request(nc, "todo.list.get_by_user", data, correlationID)
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

func UpdateListTitle(nc *nats.Conn, id, userID, title, correlationID string) (*todov1.List, error) {
	req := &todov1.UpdateListTitleRequest{
		Id:     id,
		UserId: userID,
		Title:  title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal update list title request: %w", err)
	}

	msg, err := request(nc, "todo.list.update_title", data, correlationID)
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

func DeleteList(nc *nats.Conn, id, userID, correlationID string) error {
	req := &todov1.DeleteListRequest{
		Id:     id,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal delete list request: %w", err)
	}

	msg, err := request(nc, "todo.list.delete", data, correlationID)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	// DeleteListResponse is empty on success — only check for errors
	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func CreateItem(nc *nats.Conn, listID, userID, title, correlationID string) (*todov1.Item, error) {
	req := &todov1.CreateItemRequest{
		ListId: listID,
		UserId: userID,
		Title:  title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal create item request: %w", err)
	}

	msg, err := request(nc, "todo.item.create", data, correlationID)
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

func GetAllItemsFromList(nc *nats.Conn, listID, userID, correlationID string) ([]*todov1.Item, error) {
	req := &todov1.GetAllItemsFromListRequest{
		ListId: listID,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get items request: %w", err)
	}

	msg, err := request(nc, "todo.item.get_all", data, correlationID)
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

func ToggleItemCompleted(nc *nats.Conn, id, userID, correlationID string) (*todov1.Item, error) {
	req := &todov1.ToggleItemCompletedRequest{
		Id:     id,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal toggle item request: %w", err)
	}

	msg, err := request(nc, "todo.item.toggle_completed", data, correlationID)
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

func DeleteItem(nc *nats.Conn, id, userID, listID, correlationID string) error {
	req := &todov1.DeleteItemRequest{
		Id:     id,
		UserId: userID,
		ListId: listID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal delete item request: %w", err)
	}

	msg, err := request(nc, "todo.item.delete", data, correlationID)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func AddListMember(nc *nats.Conn, listID, userID, memberEmail, correlationID string) (*todov1.ListMember, error) {
	req := &todov1.AddListMemberRequest{
		ListId:      listID,
		UserId:      userID,
		MemberEmail: memberEmail,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal add list member request: %w", err)
	}

	msg, err := request(nc, "todo.list.add_member", data, correlationID)
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

func RemoveListMember(nc *nats.Conn, listID, userID, memberUserID, correlationID string) error {
	req := &todov1.RemoveListMemberRequest{
		ListId:       listID,
		UserId:       userID,
		MemberUserId: memberUserID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal remove list member request: %w", err)
	}

	msg, err := request(nc, "todo.list.remove_member", data, correlationID)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func GetListMembers(nc *nats.Conn, listID, userID, correlationID string) ([]*todov1.ListMember, error) {
	req := &todov1.GetListMembersRequest{
		ListId: listID,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get list members request: %w", err)
	}

	msg, err := request(nc, "todo.list.get_members", data, correlationID)
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

func CreateTemplate(nc *nats.Conn, userID, title, correlationID string) (*todov1.Template, error) {
	req := &todov1.CreateTemplateRequest{
		UserId: userID,
		Title:  title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal create template request: %w", err)
	}

	msg, err := request(nc, "todo.template.create", data, correlationID)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.CreateTemplateResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.Template != nil {
		return resp.Template, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func GetTemplatesByUser(nc *nats.Conn, userID, correlationID string) ([]*todov1.Template, error) {
	req := &todov1.GetTemplatesByUserRequest{
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get templates request: %w", err)
	}

	msg, err := request(nc, "todo.template.get_by_user", data, correlationID)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.GetTemplatesByUserResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil {
		return resp.Templates, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func UpdateTemplateTitle(nc *nats.Conn, id, userID, title, correlationID string) (*todov1.Template, error) {
	req := &todov1.UpdateTemplateTitleRequest{
		Id:     id,
		UserId: userID,
		Title:  title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal update template title request: %w", err)
	}

	msg, err := request(nc, "todo.template.update_title", data, correlationID)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.UpdateTemplateTitleResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.Template != nil {
		return resp.Template, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func DeleteTemplate(nc *nats.Conn, id, userID, correlationID string) error {
	req := &todov1.DeleteTemplateRequest{
		Id:     id,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal delete template request: %w", err)
	}

	msg, err := request(nc, "todo.template.delete", data, correlationID)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	// DeleteTemplateResponse is empty on success — only check for errors
	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func CreateTemplateItem(nc *nats.Conn, templateID, userID, title, correlationID string) (*todov1.TemplateItem, error) {
	req := &todov1.CreateTemplateItemRequest{
		TemplateId: templateID,
		UserId:     userID,
		Title:      title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal create template item request: %w", err)
	}

	msg, err := request(nc, "todo.template_item.create", data, correlationID)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.CreateTemplateItemResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.Item != nil {
		return resp.Item, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func GetTemplateItems(nc *nats.Conn, templateID, userID, correlationID string) ([]*todov1.TemplateItem, error) {
	req := &todov1.GetTemplateItemsRequest{
		TemplateId: templateID,
		UserId:     userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal get template items request: %w", err)
	}

	msg, err := request(nc, "todo.template_item.get_all", data, correlationID)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.GetTemplateItemsResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil {
		return resp.Items, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}

func DeleteTemplateItem(nc *nats.Conn, id, userID, correlationID string) error {
	req := &todov1.DeleteTemplateItemRequest{
		Id:     id,
		UserId: userID,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal delete template item request: %w", err)
	}

	msg, err := request(nc, "todo.template_item.delete", data, correlationID)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}

	// DeleteTemplateItemResponse is empty on success — only check for errors
	if svcErr := parseError(msg.Data); svcErr != nil {
		return svcErr
	}

	return nil
}

func UseTemplate(nc *nats.Conn, templateID, userID, title, correlationID string) (*todov1.List, error) {
	req := &todov1.UseTemplateRequest{
		TemplateId: templateID,
		UserId:     userID,
		Title:      title,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal use template request: %w", err)
	}

	msg, err := request(nc, "todo.template.use", data, correlationID)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp todov1.UseTemplateResponse
	if err := proto.Unmarshal(msg.Data, &resp); err == nil && resp.List != nil {
		return resp.List, nil
	}

	if svcErr := parseError(msg.Data); svcErr != nil {
		return nil, svcErr
	}

	return nil, fmt.Errorf("unexpected response from todo service")
}
