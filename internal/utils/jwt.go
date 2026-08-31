package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GetJWTSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = "supersecretkey"
	}
	return []byte(secret)
}

func GetJWTRefreshSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("JWT_REFRESH_SECRET"))
	if secret == "" {
		secret = "supersecretrefreshkey"
	}
	return []byte(secret)
}

func GenerateAccessToken(userID uint, sessionID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"session_id": sessionID,
		"user_id":    userID,
		"iat":        now.Unix(),
		"exp":        now.Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecret())
}

func GenerateRefreshToken(userID uint, sessionID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"session_id": sessionID,
		"user_id":    userID,
		"iat":        now.Unix(),
		"exp":        now.Add(30 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTRefreshSecret())
}

func SignUpToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Cant able to convert bytes into random numbers")
	}
	return hex.EncodeToString(b)
}

func ReseTToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Cant able to convert bytes into random numbers")
	}
	return hex.EncodeToString(b)
}
