package main

import (
	"log"
	"net"

	"auth-microservice/internal/auth"
	"auth-microservice/internal/config"
	"auth-microservice/internal/middleware"
	"auth-microservice/internal/user"
	"auth-microservice/pkg/db"
	"auth-microservice/pkg/jwt"
	authProto "auth-microservice/proto/auth"
	userProto "auth-microservice/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg := config.New()
	log.Printf("🚀 Starting Auth Microservice...")
	log.Printf("📝 Config - Port: %s, DB: %s", cfg.Port, cfg.DBName)

	// Connect to MongoDB
	mongoDB, err := db.NewMongoDB(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	// Test MongoDB connection
	if err := mongoDB.TestConnection(); err != nil {
		log.Fatalf("❌ MongoDB connection test failed: %v", err)
	}

	// Create JWT service
	jwtService := jwt.NewJWTService(cfg.JWTSecret)
	log.Println("🔐 JWT service initialized")

	// Test JWT service
	testToken, err := jwtService.GenerateToken("test-id", "test@example.com", "user")
	if err != nil {
		log.Fatalf("❌ JWT test failed: %v", err)
	}
	log.Printf("🎫 JWT test successful - token length: %d", len(testToken))

	// Test JWT validation
	claims, err := jwtService.ValidateToken(testToken)
	if err != nil {
		log.Fatalf("❌ JWT validation test failed: %v", err)
	}
	log.Printf("✅ JWT validation successful - User: %s, Role: %s", claims.Email, claims.Role)

	// Initialize repositories
	userRepo := user.NewRepository(mongoDB)
	log.Println("📊 User repository initialized")

	// Initialize services
	authService := auth.NewService(userRepo, jwtService, mongoDB)
	log.Println("🔐 Auth service initialized")

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userRepo)
	log.Println("🎯 gRPC handlers initialized")

	// Initialize middleware
	authInterceptor := middleware.AuthInterceptor(jwtService)
	loggingInterceptor := middleware.LoggingInterceptor()

	// Create gRPC server with interceptors
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			authInterceptor,
		),
	)

	// Register services
	authProto.RegisterAuthServiceServer(server, authHandler)
	userProto.RegisterUserServiceServer(server, userHandler)
	log.Println("📡 gRPC services registered")

	// Enable reflection (สำหรับ grpcurl testing)
	reflection.Register(server)
	log.Println("🔍 gRPC reflection enabled")

	// Start listening
	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	log.Printf("✅ Auth Microservice started successfully!")
	log.Printf("🌐 gRPC server listening on port %s", cfg.Port)
	log.Printf("🍃 MongoDB connected: %s", cfg.MongoURI)
	log.Printf("📚 Collections: users, blacklisted_tokens, rate_limits")
	log.Printf("🎯 Ready for API development!")
	log.Printf("")
	log.Printf("📋 Available Services:")
	log.Printf("   🔐 AuthService: Login, Logout, Register")
	log.Printf("   👥 UserService: ListUsers, GetProfile, UpdateProfile, DeleteProfile")
	log.Printf("")
	log.Printf("🧪 Test Credentials:")
	log.Printf("   📧 Email: admin@example.com")
	log.Printf("   🔑 Password: password")

	if err := server.Serve(listener); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}
