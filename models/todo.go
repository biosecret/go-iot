package models

import "time"

// Todo là cấu trúc dữ liệu của một todo
type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Completed   bool      `json:"completed"`
	Description string    `json:"description"`
	Date        string    `json:"date"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"` // Lưu mật khẩu đã được mã hóa (hashed)
}
