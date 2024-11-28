package handlers

import (
	"os"
	"time"

	"github.com/biosecret/go-iot/database"
	"github.com/biosecret/go-iot/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler đăng ký người dùng mới
func RegisterHandler(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "could not hash password"})
	}
	user.Password = string(hashedPassword)

	// Lưu người dùng vào database
	_, err = database.GetDB().Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "user registered successfully"})
}

func LoginHandler(c *fiber.Ctx) error {
	var input models.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Kiểm tra thông tin người dùng từ database
	var user models.User
	err := database.GetDB().QueryRow("SELECT id, username, password FROM users WHERE username=$1", input.Username).
		Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	// So khớp mật khẩu
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	// Tạo access token và refresh token
	accessToken, err := generateJWT(user.ID, "15m")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	refreshToken, err := generateJWT(user.ID, "7d")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Tạo JWT token
func generateJWT(userID int64, duration string) (string, error) {
	// Chuyển đổi duration (ví dụ: "7d" -> 7 * 24 giờ)
	var expirationTime time.Duration
	if duration == "7d" {
		expirationTime = 7 * 24 * time.Hour // 7 ngày
	} else {
		expirationTime, _ = time.ParseDuration(duration) // Các duration hợp lệ khác
	}

	// Tạo claims cho JWT
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expirationTime).Unix(),
	}

	// Tạo token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
