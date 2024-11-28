package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomID tạo một ID ngẫu nhiên có độ dài 16 byte (32 ký tự)
func GenerateRandomID() (string, error) {
	bytes := make([]byte, 16) // 16 byte -> 32 ký tự hex
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
