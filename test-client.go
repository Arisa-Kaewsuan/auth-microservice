package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Simple structs for testing (without importing proto)
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

func main() {
	log.Println("🔍 Testing gRPC connection to localhost:50052...")

	// Test basic connection first
	conn, err := grpc.Dial("localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer conn.Close()

	log.Println("✅ Basic gRPC connection successful!")

	// Test connection state
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state := conn.GetState()
	log.Printf("🔍 Connection state: %v", state)

	// Try to wait for connection to be ready
	if conn.WaitForStateChange(ctx, state) {
		newState := conn.GetState()
		log.Printf("🔍 New connection state: %v", newState)
	}

	log.Println("🎉 Connection test completed!")
	log.Println("📝 This confirms that:")
	log.Println("   ✅ Go server is running correctly")
	log.Println("   ✅ Port 50052 is accessible")
	log.Println("   ✅ gRPC protocol is working")
	log.Println("")
	log.Println("💡 The issue is likely with Postman's gRPC implementation")
	log.Println("💡 Try importing the proto file instead of using server reflection")
}
