package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	commonv1 "github.com/terjelafton/platform/libs/proto-stubs/common/v1"
	idv1 "github.com/terjelafton/platform/libs/proto-stubs/id/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("failed to connect to NATS:", err)
	}
	defer nc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("=== Testing ID Service ===\n")

	// Test 1: Register new user
	fmt.Println("1. Registering new user...")
	registerReq := &idv1.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	registerData, _ := proto.Marshal(registerReq)

	msg, err := nc.RequestWithContext(ctx, "id.auth.register", registerData)
	if err != nil {
		log.Fatal("register request failed:", err)
	}

	var registerResp idv1.RegisterResponse
	if err := proto.Unmarshal(msg.Data, &registerResp); err != nil {
		var errResp commonv1.ErrorResponse
		if err := proto.Unmarshal(msg.Data, &errResp); err == nil {
			fmt.Printf("❌ Registration failed: %s - %s\n", errResp.Code, errResp.Message)
		} else {
			log.Fatal("failed to unmarshal response:", err)
		}
	} else {
		fmt.Printf("✓ User registered: %s (%s)\n", registerResp.User.Email, registerResp.User.Id)
		fmt.Printf("✓ Token received: %s...\n\n", registerResp.Token[:20])
	}

	// Test 2: Login with same credentials
	fmt.Println("2. Logging in with same credentials...")
	loginReq := &idv1.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	loginData, _ := proto.Marshal(loginReq)

	msg, err = nc.RequestWithContext(ctx, "id.auth.login", loginData)
	if err != nil {
		log.Fatal("login request failed:", err)
	}

	var loginResp idv1.LoginResponse
	var token string
	if err := proto.Unmarshal(msg.Data, &loginResp); err != nil {
		var errResp commonv1.ErrorResponse
		if err := proto.Unmarshal(msg.Data, &errResp); err == nil {
			fmt.Printf("❌ Login failed: %s - %s\n", errResp.Code, errResp.Message)
		} else {
			log.Fatal("failed to unmarshal response:", err)
		}
	} else {
		token = loginResp.Token
		fmt.Printf("✓ Login successful: %s\n", loginResp.User.Email)
		fmt.Printf("✓ Token received: %s...\n\n", token[:20])
	}

	// Test 3: Validate token
	if token != "" {
		fmt.Println("3. Validating token...")
		validateReq := &idv1.ValidateTokenRequest{
			Token: token,
		}
		validateData, _ := proto.Marshal(validateReq)

		msg, err = nc.RequestWithContext(ctx, "id.auth.validate", validateData)
		if err != nil {
			log.Fatal("validate request failed:", err)
		}

		var validateResp idv1.ValidateTokenResponse
		if err := proto.Unmarshal(msg.Data, &validateResp); err != nil {
			var errResp commonv1.ErrorResponse
			if err := proto.Unmarshal(msg.Data, &errResp); err == nil {
				fmt.Printf("❌ Validation failed: %s - %s\n", errResp.Code, errResp.Message)
			} else {
				log.Fatal("failed to unmarshal response:", err)
			}
		} else {
			fmt.Printf("✓ Token valid for user: %s\n\n", validateResp.UserId)
		}
	}

	// Test 4: Try to login with wrong password
	fmt.Println("4. Testing wrong password...")
	wrongLoginReq := &idv1.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	wrongLoginData, _ := proto.Marshal(wrongLoginReq)

	msg, err = nc.RequestWithContext(ctx, "id.auth.login", wrongLoginData)
	if err != nil {
		log.Fatal("login request failed:", err)
	}

	var wrongLoginResp idv1.LoginResponse
	if err := proto.Unmarshal(msg.Data, &wrongLoginResp); err != nil {
		var errResp commonv1.ErrorResponse
		if err := proto.Unmarshal(msg.Data, &errResp); err == nil {
			fmt.Printf("✓ Correctly rejected: %s - %s\n", errResp.Code, errResp.Message)
		}
	} else {
		fmt.Println("❌ Should have failed with wrong password!")
	}

	fmt.Println("\n=== Done ===")
}
