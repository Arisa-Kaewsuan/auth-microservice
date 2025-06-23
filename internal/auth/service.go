package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"auth-microservice/internal/models"
	"auth-microservice/internal/user"
	"auth-microservice/pkg/db"
	"auth-microservice/pkg/jwt"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/time/rate"
)

type Service struct {
	userRepo   *user.Repository
	jwtService *jwt.JWTService
	db         *db.MongoDB
	limiter    *rate.Limiter
}

func NewService(userRepo *user.Repository, jwtService *jwt.JWTService, database *db.MongoDB) *Service {
	// Rate limiter: 5 attempts per minute
	limiter := rate.NewLimiter(rate.Every(12*time.Second), 5)

	return &Service{
		userRepo:   userRepo,
		jwtService: jwtService,
		db:         database,
		limiter:    limiter,
	}
}

// Login - เข้าสู่ระบบ
func (s *Service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// Rate limiting check
	if !s.limiter.Allow() {
		log.Printf("Rate limit exceeded for login attempt")
		return &LoginResponse{
			Success: false,
			Message: "Too many login attempts. Please try again later.",
		}, nil
	}

	// Validate input
	if email == "" || password == "" {
		return &LoginResponse{
			Success: false,
			Message: "Email and password are required",
		}, nil
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("Login failed - user not found: %s", email)
		return &LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// Check password
	if !user.CheckPassword(password) {
		log.Printf("Login failed - invalid password: %s", email)
		return &LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		log.Printf("Failed to generate token for user %s: %v", email, err)
		return &LoginResponse{
			Success: false,
			Message: "Internal server error",
		}, fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("Login successful for user: %s", email)

	return &LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		User:    user.ToSafeUser(),
	}, nil
}

// Logout - ออกจากระบบ
func (s *Service) Logout(ctx context.Context, token string) (*LogoutResponse, error) {
	// Validate token first
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return &LogoutResponse{
			Success: false,
			Message: "Invalid token",
		}, nil
	}

	// Add token to blacklist
	blacklistedToken := bson.M{
		"token":      token,
		"expires_at": claims.ExpiresAt.Time,
		"created_at": time.Now(),
	}

	_, err = s.db.BlacklistedTokens().InsertOne(ctx, blacklistedToken)
	if err != nil {
		log.Printf("Failed to blacklist token: %v", err)
		return &LogoutResponse{
			Success: false,
			Message: "Internal server error",
		}, fmt.Errorf("failed to blacklist token: %w", err)
	}

	log.Printf("User logged out successfully: %s", claims.Email)

	return &LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

// Register - สมัครสมาชิก
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Validate input
	if err := s.validateRegisterRequest(req); err != nil {
		return &RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Check if email already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return &RegisterResponse{
			Success: false,
			Message: "Email already registered",
		}, nil
	}

	// Create new user
	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "user", // Default role
	}

	// Hash password
	if err := user.HashPassword(req.Password); err != nil {
		log.Printf("Failed to hash password: %v", err)
		return &RegisterResponse{
			Success: false,
			Message: "Internal server error",
		}, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user
	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Printf("Failed to create user: %v", err)
		return &RegisterResponse{
			Success: false,
			Message: "Failed to create account",
		}, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User registered successfully: %s", req.Email)

	return &RegisterResponse{
		Success: true,
		Message: "Account created successfully",
		UserID:  user.ID.Hex(),
	}, nil
}

// validateRegisterRequest - ตรวจสอบข้อมูลการสมัคร
func (s *Service) validateRegisterRequest(req *RegisterRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}

	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}

	// Simple email validation
	if !isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// isValidEmail - ตรวจสอบรูปแบบ email
func isValidEmail(email string) bool {
	// Simple email validation (สำหรับ demo)
	// ในการใช้งานจริงควรใช้ regex ที่สมบูรณ์กว่า
	return len(email) > 3 &&
		contains(email, "@") &&
		contains(email, ".")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Response structs
type LoginResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Token   string                 `json:"token,omitempty"`
	User    map[string]interface{} `json:"user,omitempty"`
}

type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
