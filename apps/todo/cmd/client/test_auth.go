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

// This test verifies that authorization is working correctly
// by attempting to access another user's resources
func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Create two different users
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	fmt.Printf("👤 User 1: %s\n", userID1)
	fmt.Printf("👤 User 2: %s\n\n", userID2)

	// User 1 creates a list
	fmt.Println("User 1 creates a list...")
	createReq := &todov1.CreateListRequest{
		UserId: userID1,
		Title:  "User 1's Private List",
	}

	reqData, _ := proto.Marshal(createReq)
	msg, err := nc.Request("todo.list.create", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	var createResp todov1.CreateListResponse
	if err := proto.Unmarshal(msg.Data, &createResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	listID := createResp.List.Id
	fmt.Printf("✅ User 1 created list: %s (ID: %s)\n\n", createResp.List.Title, listID)

	// User 1 creates an item
	fmt.Println("User 1 creates an item...")
	createItemReq := &todov1.CreateItemRequest{
		ListId: listID,
		Title:  "User 1's Private Item",
		UserId: userID1,
	}

	reqData, _ = proto.Marshal(createItemReq)
	msg, err = nc.Request("todo.item.create", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	var createItemResp todov1.CreateItemResponse
	if err := proto.Unmarshal(msg.Data, &createItemResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	itemID := createItemResp.Item.Id
	fmt.Printf("✅ User 1 created item: %s (ID: %s)\n\n", createItemResp.Item.Title, itemID)

	// Now User 2 tries to access User 1's list
	fmt.Println("🚨 Testing authorization: User 2 tries to update User 1's list...")
	updateReq := &todov1.UpdateListTitleRequest{
		Id:     listID,
		UserId: userID2, // Wrong user!
		Title:  "Hacked Title",
	}

	reqData, _ = proto.Marshal(updateReq)
	msg, err = nc.Request("todo.list.update_title", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	var errResp commonv1.ErrorResponse
	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("✅ BLOCKED! Error [%s]: %s\n\n", errResp.Code, errResp.Message)
	} else {
		fmt.Println("❌ SECURITY ISSUE: User 2 was able to update User 1's list!")
		return
	}

	// User 2 tries to add an item to User 1's list
	fmt.Println("🚨 Testing authorization: User 2 tries to add item to User 1's list...")
	createItemReq2 := &todov1.CreateItemRequest{
		ListId: listID,
		Title:  "Malicious Item",
		UserId: userID2, // Wrong user!
	}

	reqData, _ = proto.Marshal(createItemReq2)
	msg, err = nc.Request("todo.item.create", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("✅ BLOCKED! Error [%s]: %s\n\n", errResp.Code, errResp.Message)
	} else {
		fmt.Println("❌ SECURITY ISSUE: User 2 was able to add item to User 1's list!")
		return
	}

	// User 2 tries to view User 1's items
	fmt.Println("🚨 Testing authorization: User 2 tries to view User 1's items...")
	getItemsReq := &todov1.GetAllItemsFromListRequest{
		ListId: listID,
		UserId: userID2, // Wrong user!
	}

	reqData, _ = proto.Marshal(getItemsReq)
	msg, err = nc.Request("todo.item.get_all", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	var getItemsResp todov1.GetAllItemsFromListResponse
	if err := proto.Unmarshal(msg.Data, &getItemsResp); err != nil {
		log.Fatal("failed to unmarshal response:", err)
	}

	if len(getItemsResp.Items) == 0 {
		fmt.Printf("✅ BLOCKED! User 2 cannot see User 1's items (returned 0 items)\n\n")
	} else {
		fmt.Printf("❌ SECURITY ISSUE: User 2 can see %d items from User 1's list!\n", len(getItemsResp.Items))
		return
	}

	// User 2 tries to toggle User 1's item
	fmt.Println("🚨 Testing authorization: User 2 tries to toggle User 1's item...")
	toggleReq := &todov1.ToggleItemCompletedRequest{
		Id:     itemID,
		UserId: userID2, // Wrong user!
	}

	reqData, _ = proto.Marshal(toggleReq)
	msg, err = nc.Request("todo.item.toggle_completed", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("✅ BLOCKED! Error [%s]: %s\n\n", errResp.Code, errResp.Message)
	} else {
		fmt.Println("❌ SECURITY ISSUE: User 2 was able to toggle User 1's item!")
		return
	}

	// User 2 tries to delete User 1's item
	fmt.Println("🚨 Testing authorization: User 2 tries to delete User 1's item...")
	deleteItemReq := &todov1.DeleteItemRequest{
		Id:     itemID,
		UserId: userID2, // Wrong user!
	}

	reqData, _ = proto.Marshal(deleteItemReq)
	msg, err = nc.Request("todo.item.delete", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("✅ BLOCKED! Error [%s]: %s\n\n", errResp.Code, errResp.Message)
	} else {
		fmt.Println("❌ SECURITY ISSUE: User 2 was able to delete User 1's item!")
		return
	}

	// User 2 tries to delete User 1's list
	fmt.Println("🚨 Testing authorization: User 2 tries to delete User 1's list...")
	deleteListReq := &todov1.DeleteListRequest{
		Id:     listID,
		UserId: userID2, // Wrong user!
	}

	reqData, _ = proto.Marshal(deleteListReq)
	msg, err = nc.Request("todo.list.delete", reqData, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	if err := proto.Unmarshal(msg.Data, &errResp); err == nil && errResp.Code != "" {
		fmt.Printf("✅ BLOCKED! Error [%s]: %s\n\n", errResp.Code, errResp.Message)
	} else {
		fmt.Println("❌ SECURITY ISSUE: User 2 was able to delete User 1's list!")
		return
	}

	// Cleanup: User 1 deletes their own resources
	fmt.Println("Cleanup: User 1 deletes their own item...")
	deleteItemReq.UserId = userID1
	reqData, _ = proto.Marshal(deleteItemReq)
	nc.Request("todo.item.delete", reqData, 5*time.Second)

	fmt.Println("Cleanup: User 1 deletes their own list...")
	deleteListReq.UserId = userID1
	reqData, _ = proto.Marshal(deleteListReq)
	nc.Request("todo.list.delete", reqData, 5*time.Second)

	fmt.Println("\n🎉 Authorization tests passed! All unauthorized access attempts were blocked.")
}
