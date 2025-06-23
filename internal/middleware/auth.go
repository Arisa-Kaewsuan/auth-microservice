package middleware

import (
	"context"
	"log"
	"strings"

	"auth-microservice/pkg/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor - Middleware สำหรับตรวจสอบ JWT
func AuthInterceptor(jwtService *jwt.JWTService) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// ข้าม auth สำหรับ APIs เหล่านี้
		publicMethods := map[string]bool{
			"/auth.AuthService/Login":    true,
			"/auth.AuthService/Register": true,
		}

		if publicMethods[info.FullMethod] {
			log.Printf("🟢 Public method accessed: %s", info.FullMethod)
			return handler(ctx, req)
		}

		// ตรวจสอบ Authorization header
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Printf("❌ No metadata found for method: %s", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "Missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			log.Printf("❌ No authorization header for method: %s", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "Missing authorization header")
		}

		// แยก Bearer token
		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == authHeader[0] {
			log.Printf("❌ Invalid authorization format for method: %s", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "Invalid authorization format")
		}

		// ตรวจสอบ JWT token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			log.Printf("❌ Invalid token for method: %s - Error: %v", info.FullMethod, err)
			return nil, status.Errorf(codes.Unauthenticated, "Invalid token")
		}

		// เพิ่ม user info ใน context
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_role", claims.Role)

		log.Printf("🟢 Authenticated user: %s (%s) for method: %s", claims.Email, claims.Role, info.FullMethod)

		return handler(ctx, req)
	}
}

// LoggingInterceptor - Middleware สำหรับ logging
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Printf("📡 gRPC call: %s", info.FullMethod)

		resp, err := handler(ctx, req)

		if err != nil {
			log.Printf("❌ gRPC error for %s: %v", info.FullMethod, err)
		} else {
			log.Printf("✅ gRPC success for %s", info.FullMethod)
		}

		return resp, err
	}
}
