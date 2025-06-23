package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	// Connection options with pooling
	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100). // สำหรับ 1000 concurrent users
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)

	log.Println("✅ Connected to MongoDB successfully!")

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close() error {
	return m.Client.Disconnect(context.Background())
}

// Collection helpers
func (m *MongoDB) Users() *mongo.Collection {
	return m.Database.Collection("users")
}

func (m *MongoDB) BlacklistedTokens() *mongo.Collection {
	return m.Database.Collection("blacklisted_tokens")
}

func (m *MongoDB) RateLimits() *mongo.Collection {
	return m.Database.Collection("rate_limits")
}

// TestConnection - ทดสอบการเชื่อมต่อและข้อมูล
func (m *MongoDB) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test 1: List collections
	collections, err := m.Database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}
	log.Printf("📋 Available collections: %v", collections)

	// Test 2: Count users
	userCount, err := m.Users().CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	log.Printf("📊 Total users in database: %d", userCount)

	// Test 3: Find admin user
	var adminUser bson.M
	err = m.Users().FindOne(ctx, bson.M{"role": "admin"}).Decode(&adminUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("⚠️  No admin user found - consider creating one for testing")
		} else {
			return fmt.Errorf("failed to find admin user: %w", err)
		}
	} else {
		if email, ok := adminUser["email"].(string); ok {
			log.Printf("👤 Found admin user: %s", email)
		}
	}

	// Test 4: Test indexes
	indexes, err := m.Users().Indexes().List(ctx)
	if err != nil {
		log.Printf("⚠️  Warning: Could not list indexes: %v", err)
	} else {
		indexCount := 0
		for indexes.Next(ctx) {
			indexCount++
		}
		log.Printf("🔍 Users collection has %d indexes", indexCount)
	}

	// Test 5: Performance test - simple query
	start := time.Now()
	_, err = m.Users().FindOne(ctx, bson.M{"role": "admin"}).DecodeBytes()
	duration := time.Since(start)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("⚠️  Query performance test failed: %v", err)
	} else {
		log.Printf("⚡ Query performance: %v", duration)
	}

	return nil
}
