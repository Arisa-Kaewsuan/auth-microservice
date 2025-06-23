package auth

import (
	"context"
	"log"

	"auth-microservice/proto/auth"
)

// Handler - gRPC handler สำหรับ Authentication
type Handler struct {
	auth.UnimplementedAuthServiceServer
	service *Service
}

// NewHandler - สร้าง auth handler ใหม่
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Login - gRPC handler สำหรับ login
func (h *Handler) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	log.Printf("🔐 Login request received for email: %s", req.Email)

	// เรียก service layer
	result, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		log.Printf("❌ Login service error: %v", err)
		return &auth.LoginResponse{
			Success: false,
			Message: "Internal server error",
		}, nil
	}

	// แปลง response
	response := &auth.LoginResponse{
		Success: result.Success,
		Message: result.Message,
		Token:   result.Token,
	}

	if result.Success {
		log.Printf("✅ Login successful for: %s", req.Email)
	} else {
		log.Printf("❌ Login failed for: %s - %s", req.Email, result.Message)
	}

	return response, nil
}

// Logout - gRPC handler สำหรับ logout
func (h *Handler) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	log.Printf("🚪 Logout request received")

	// เรียก service layer
	result, err := h.service.Logout(ctx, req.Token)
	if err != nil {
		log.Printf("❌ Logout service error: %v", err)
		return &auth.LogoutResponse{
			Success: false,
			Message: "Internal server error",
		}, nil
	}

	// แปลง response
	response := &auth.LogoutResponse{
		Success: result.Success,
		Message: result.Message,
	}

	if result.Success {
		log.Printf("✅ Logout successful")
	} else {
		log.Printf("❌ Logout failed: %s", result.Message)
	}

	return response, nil
}

// Register - gRPC handler สำหรับ register
func (h *Handler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	log.Printf("📝 Register request received for email: %s", req.Email)

	// แปลง gRPC request เป็น service request
	serviceReq := &RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// เรียก service layer
	result, err := h.service.Register(ctx, serviceReq)
	if err != nil {
		log.Printf("❌ Register service error: %v", err)
		return &auth.RegisterResponse{
			Success: false,
			Message: "Internal server error",
		}, nil
	}

	// แปลง response
	response := &auth.RegisterResponse{
		Success: result.Success,
		Message: result.Message,
		UserId:  result.UserID,
	}

	if result.Success {
		log.Printf("✅ Registration successful for: %s", req.Email)
	} else {
		log.Printf("❌ Registration failed for: %s - %s", req.Email, result.Message)
	}

	return response, nil
}
