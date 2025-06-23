package main

import (
	"log"
	"net"

	"auth-microservice/internal/config"
	"auth-microservice/pkg/db"
	"auth-microservice/pkg/jwt"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg := config.New()

	// Connect to MongoDB
	mongoDB, err := db.NewMongoDB(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	// Create JWT service
	jwtService := jwt.NewJWTService(cfg.JWTSecret)

	// Create gRPC server
	server := grpc.NewServer()

	// TODO: Register services here
	// auth.RegisterAuthServiceServer(server, authService)
	// user.RegisterUserServiceServer(server, userService)

	// Start listening
	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Auth Microservice started on port %s", cfg.Port)
	log.Printf("MongoDB connected to: %s", cfg.MongoURI)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
