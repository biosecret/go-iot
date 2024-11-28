package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware xác thực access token
func JWTMiddleware(c *fiber.Ctx) error {
	// Lấy token từ header Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "missing token"})
	}

	// Tách từ "Bearer <token>"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return c.Status(401).JSON(fiber.Map{"error": "invalid token format"})
	}

	// Parse và kiểm tra token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid or expired token"})
	}

	// Lưu thông tin user ID vào context nếu token hợp lệ
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		c.Locals("user_id", claims["user_id"])
	}

	return c.Next()
}
