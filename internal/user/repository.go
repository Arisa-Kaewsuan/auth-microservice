package user

import (
	"context"
	"fmt"
	"time"

	"auth-microservice/internal/models"
	"auth-microservice/pkg/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	db *db.MongoDB
}

func NewRepository(database *db.MongoDB) *Repository {
	return &Repository{
		db: database,
	}
}

// GetByEmail - หา user ตาม email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	filter := bson.M{
		"email":      email,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}

	err := r.db.Users().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// GetByID - หา user ตาม ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user models.User
	filter := bson.M{
		"_id":        objectID,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}

	err = r.db.Users().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// Create - สร้าง user ใหม่
func (r *Repository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true

	_, err := r.db.Users().InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("email already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Update - อัพเดท user
func (r *Repository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"updated_at": user.UpdatedAt,
		},
	}

	result, err := r.db.Users().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// SoftDelete - ลบ user แบบ soft delete
func (r *Repository) SoftDelete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	now := time.Now()
	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"is_active":  false,
			"updated_at": now,
		},
	}

	result, err := r.db.Users().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List - แสดงรายการ users พร้อม filtering และ pagination
func (r *Repository) List(ctx context.Context, nameFilter, emailFilter string, skip, limit int) ([]*models.User, int64, error) {
	// Build filter
	filter := bson.M{
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}

	if nameFilter != "" {
		filter["$or"] = []bson.M{
			{"first_name": bson.M{"$regex": nameFilter, "$options": "i"}},
			{"last_name": bson.M{"$regex": nameFilter, "$options": "i"}},
		}
	}

	if emailFilter != "" {
		filter["email"] = bson.M{"$regex": emailFilter, "$options": "i"}
	}

	// Count total
	total, err := r.db.Users().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Find with pagination
	skip64 := int64(skip)
	limit64 := int64(limit)
	cursor, err := r.db.Users().Find(ctx, filter, &options.FindOptions{
		Skip:  &skip64,
		Limit: &limit64,
		Sort:  bson.D{{"created_at", -1}}, // เรียงจากใหม่ไปเก่า
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, 0, fmt.Errorf("failed to decode user: %w", err)
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, fmt.Errorf("cursor error: %w", err)
	}

	return users, total, nil
}
