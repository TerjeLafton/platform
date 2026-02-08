package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/terjelafton/platform/libs/proto-stubs/common/v1"
	todov1 "github.com/terjelafton/platform/libs/proto-stubs/todo/v1"
)

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	userID := uuid.New().String()

	// Test CreateList
	fmt.Println("Testing CreateList...")
	createReq := &todov1.CreateListRequest{
		UserId: userID,
		Title:  "Shopping",
	}

	reqData, _ := proto.Marshal(createReq)
	msg, err := nc.Request("todo.list.create", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	var errResp commonv1.ErrorResponse
	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	var createResp todov1.CreateListResponse
	if err := proto.Unmarshal(msg.Data, &createResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	fmt.Printf("✅ Created list: %s (ID: %s)\n\n", createResp.List.Title, createResp.List.Id)

	// Test GetListsByUser
	fmt.Println("Testing GetListsByUser...")
	getReq := &todov1.GetListsByUserRequest{
		UserId: userID,
	}

	reqData, _ = proto.Marshal(getReq)
	msg, err = nc.Request("todo.list.get_by_user", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	var getResp todov1.GetListsByUserResponse
	if err := proto.Unmarshal(msg.Data, &getResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	fmt.Printf("✅ Retrieved %d list(s):\n", len(getResp.Lists))
	for _, list := range getResp.Lists {
		fmt.Printf("  - %s (ID: %s)\n", list.Title, list.Id)
	}

	if len(getResp.Lists) == 0 {
		fmt.Println("No lists to update")
		return
	}

	// Test UpdateListTitle
	fmt.Println("\nTesting UpdateListTitle...")
	updateReq := &todov1.UpdateListTitleRequest{
		Id:     createResp.List.Id,
		UserId: userID,
		Title:  "Updated Shopping List",
	}

	reqData, _ = proto.Marshal(updateReq)
	msg, err = nc.Request("todo.list.update_title", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	var updateResp todov1.UpdateListTitleResponse
	if err := proto.Unmarshal(msg.Data, &updateResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	fmt.Printf("✅ Updated list title: %s (ID: %s)\n", updateResp.List.Title, updateResp.List.Id)

	// Test CreateItem (create 2 items)
	fmt.Println("\nTesting CreateItem...")
	createItemReq1 := &todov1.CreateItemRequest{
		ListId: createResp.List.Id,
		Title:  "Buy milk",
		UserId: userID,
	}

	reqData, _ = proto.Marshal(createItemReq1)
	msg, err = nc.Request("todo.item.create", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	var createItemResp1 todov1.CreateItemResponse
	if err := proto.Unmarshal(msg.Data, &createItemResp1); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	fmt.Printf("✅ Created item: %s (ID: %s)\n", createItemResp1.Item.Title, createItemResp1.Item.Id)

	createItemReq2 := &todov1.CreateItemRequest{
		ListId: createResp.List.Id,
		Title:  "Buy bread",
		UserId: userID,
	}

	reqData, _ = proto.Marshal(createItemReq2)
	msg, err = nc.Request("todo.item.create", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	var createItemResp2 todov1.CreateItemResponse
	if err := proto.Unmarshal(msg.Data, &createItemResp2); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	fmt.Printf("✅ Created item: %s (ID: %s)\n", createItemResp2.Item.Title, createItemResp2.Item.Id)

	// Test GetAllItemsFromList
	fmt.Println("\nTesting GetAllItemsFromList...")
	getAllItemsReq := &todov1.GetAllItemsFromListRequest{
		ListId: createResp.List.Id,
		UserId: userID,
	}

	reqData, _ = proto.Marshal(getAllItemsReq)
	msg, err = nc.Request("todo.item.get_all", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	var getAllItemsResp todov1.GetAllItemsFromListResponse
	if err := proto.Unmarshal(msg.Data, &getAllItemsResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	fmt.Printf("✅ Retrieved %d item(s):\n", len(getAllItemsResp.Items))
	for _, item := range getAllItemsResp.Items {
		fmt.Printf("  - %s (ID: %s, Completed: %v)\n", item.Title, item.Id, item.Completed)
	}

	// Test ToggleItemCompleted
	fmt.Println("\nTesting ToggleItemCompleted...")
	toggleReq := &todov1.ToggleItemCompletedRequest{
		Id:     createItemResp1.Item.Id,
		UserId: userID,
	}

	reqData, _ = proto.Marshal(toggleReq)
	msg, err = nc.Request("todo.item.toggle_completed", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	var toggleResp todov1.ToggleItemCompletedResponse
	if err := proto.Unmarshal(msg.Data, &toggleResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	fmt.Printf("✅ Toggled item: %s (Completed: %v)\n", toggleResp.Item.Title, toggleResp.Item.Completed)

	// Test DeleteItem
	fmt.Println("\nTesting DeleteItem...")
	deleteItemReq := &todov1.DeleteItemRequest{
		Id:     createItemResp2.Item.Id,
		UserId: userID,
	}

	reqData, _ = proto.Marshal(deleteItemReq)
	msg, err = nc.Request("todo.item.delete", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	fmt.Printf("✅ Deleted item (ID: %s)\n", createItemResp2.Item.Id)

	// Test GetAllItemsFromList again (should have 1 item left)
	fmt.Println("\nTesting GetAllItemsFromList (after delete)...")
	reqData, _ = proto.Marshal(getAllItemsReq)
	msg, err = nc.Request("todo.item.get_all", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	if err := proto.Unmarshal(msg.Data, &getAllItemsResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	fmt.Printf("✅ Retrieved %d item(s) after delete:\n", len(getAllItemsResp.Items))
	for _, item := range getAllItemsResp.Items {
		fmt.Printf("  - %s (ID: %s, Completed: %v)\n", item.Title, item.Id, item.Completed)
	}

	// Test DeleteList (cleanup)
	fmt.Println("\nTesting DeleteList (cleanup)...")
	deleteReq := &todov1.DeleteListRequest{
		Id:     createResp.List.Id,
		UserId: userID,
	}

	reqData, _ = proto.Marshal(deleteReq)
	msg, err = nc.Request("todo.list.delete", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("❌ Error [%s]: %s\n", errResp.Code, errResp.Message)
		return
	}

	fmt.Printf("✅ Deleted list (ID: %s)\n", createResp.List.Id)
	fmt.Println("\n🎉 All tests passed!")
}
