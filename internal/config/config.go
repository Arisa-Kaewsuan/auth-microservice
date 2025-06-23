package config

import (
	"os"
)

type Config struct {
	Port      string
	MongoURI  string
	DBName    string
	JWTSecret string
}

func New() *Config {
	return &Config{
		Port:      getEnv("PORT", "50051"),
		MongoURI:  getEnv("MONGO_URI", "mongodb://admin:password123@localhost:27017"),
		DBName:    getEnv("DB_NAME", "auth_microservice"),
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
