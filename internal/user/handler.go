package user

import (
	"context"
	"log"

	"auth-microservice/proto/user"
)

// Handler - gRPC handler สำหรับ User Management
type Handler struct {
	user.UnimplementedUserServiceServer
	repository *Repository
}

// NewHandler - สร้าง user handler ใหม่
func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

// ListUsers - gRPC handler สำหรับแสดงรายการ users
func (h *Handler) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	log.Printf("📋 ListUsers request - Page: %d, Limit: %d", req.Page, req.Limit)

	// คำนวณ pagination
	page := int(req.Page)
	limit := int(req.Limit)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	skip := (page - 1) * limit

	// เรียก repository
	users, total, err := h.repository.List(ctx, req.NameFilter, req.EmailFilter, skip, limit)
	if err != nil {
		log.Printf("❌ ListUsers repository error: %v", err)
		return &user.ListUsersResponse{
			Users:      []*user.User{},
			Total:      0,
			Page:       int32(page),
			Limit:      int32(limit),
			TotalPages: 0,
		}, nil
	}

	// แปลง users เป็น proto format
	protoUsers := make([]*user.User, 0, len(users))
	for _, u := range users {
		protoUser := &user.User{
			Id:        u.ID.Hex(),
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Role:      u.Role,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		protoUsers = append(protoUsers, protoUser)
	}

	// คำนวณ total pages
	totalPages := (int(total) + limit - 1) / limit

	response := &user.ListUsersResponse{
		Users:      protoUsers,
		Total:      int32(total),
		Page:       int32(page),
		Limit:      int32(limit),
		TotalPages: int32(totalPages),
	}

	log.Printf("✅ ListUsers successful - Found %d users", len(protoUsers))
	return response, nil
}

// GetProfile - gRPC handler สำหรับดึงข้อมูล user profile
func (h *Handler) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, error) {
	log.Printf("👤 GetProfile request for ID: %s", req.UserId)

	// หา user ตาม ID
	userData, err := h.repository.GetByID(ctx, req.UserId)
	if err != nil {
		log.Printf("❌ GetProfile repository error: %v", err)
		return &user.GetProfileResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	// แปลงเป็น proto format
	protoUser := &user.User{
		Id:        userData.ID.Hex(),
		Email:     userData.Email,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Role:      userData.Role,
		CreatedAt: userData.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response := &user.GetProfileResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		User:    protoUser,
	}

	log.Printf("✅ GetProfile successful for: %s", userData.Email)
	return response, nil
}

// UpdateProfile - gRPC handler สำหรับอัพเดท user profile
func (h *Handler) UpdateProfile(ctx context.Context, req *user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	log.Printf("✏️ UpdateProfile request for ID: %s", req.UserId)

	// หา user ที่จะอัพเดท
	userData, err := h.repository.GetByID(ctx, req.UserId)
	if err != nil {
		log.Printf("❌ UpdateProfile - user not found: %v", err)
		return &user.UpdateProfileResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	// อัพเดทข้อมูล
	if req.FirstName != "" {
		userData.FirstName = req.FirstName
	}
	if req.LastName != "" {
		userData.LastName = req.LastName
	}
	if req.Email != "" {
		userData.Email = req.Email
	}

	// บันทึกการเปลี่ยนแปลง
	err = h.repository.Update(ctx, userData)
	if err != nil {
		log.Printf("❌ UpdateProfile repository error: %v", err)
		return &user.UpdateProfileResponse{
			Success: false,
			Message: "Failed to update profile",
		}, nil
	}

	response := &user.UpdateProfileResponse{
		Success: true,
		Message: "Profile updated successfully",
	}

	log.Printf("✅ UpdateProfile successful for: %s", userData.Email)
	return response, nil
}

// DeleteProfile - gRPC handler สำหรับลบ user profile (soft delete)
func (h *Handler) DeleteProfile(ctx context.Context, req *user.DeleteProfileRequest) (*user.DeleteProfileResponse, error) {
	log.Printf("🗑️ DeleteProfile request for ID: %s", req.UserId)

	// ลบ user (soft delete)
	err := h.repository.SoftDelete(ctx, req.UserId)
	if err != nil {
		log.Printf("❌ DeleteProfile repository error: %v", err)
		return &user.DeleteProfileResponse{
			Success: false,
			Message: "Failed to delete profile",
		}, nil
	}

	response := &user.DeleteProfileResponse{
		Success: true,
		Message: "Profile deleted successfully",
	}

	log.Printf("✅ DeleteProfile successful for ID: %s", req.UserId)
	return response, nil
}
