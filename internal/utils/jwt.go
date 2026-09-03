package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
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

func RefreshAccessToken(refreshToken string) (string, error) {
	token, err := jwt.Parse(
		refreshToken,
		func(token *jwt.Token) (interface{}, error) {
			return GetJWTRefreshSecret(), nil
		},
	)

	if err != nil || !token.Valid {
		return "", errors.New("invalid or expired refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return "", errors.New("invalid user id")
	}

	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return "", errors.New("invalid session id")
	}

	// Only generate a NEW ACCESS TOKEN
	return GenerateAccessToken(
		uint(userIDFloat),
		sessionID,
	)
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
