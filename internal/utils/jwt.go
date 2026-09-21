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

// Get the secret key for access tokens.
func GetJWTSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	if secret == "" {
		secret = "supersecretkey"
	}

	return []byte(secret)
}

// Get the secret key for refresh tokens.
func GetJWTRefreshSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("JWT_REFRESH_SECRET"))

	if secret == "" {
		secret = "supersecretrefreshkey"
	}

	return []byte(secret)
}

// Generate a new access token.
func GenerateAccessToken(userID uint, sessionID string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"iat":        now.Unix(),
		"exp":        now.Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(GetJWTSecret())
}

// Generate a new refresh token.
func GenerateRefreshToken(userID uint, sessionID string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"iat":        now.Unix(),
		"exp":        now.Add(30 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(GetJWTRefreshSecret())
}

func RefreshTokens(refreshToken string) (string, string, error) {

	token, err := jwt.Parse(
		refreshToken,
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return GetJWTRefreshSecret(), nil
		},
	)


	if err != nil || !token.Valid {
		return "", "", errors.New("invalid or expired refresh token")
	}


	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return "", "", errors.New("invalid user id")
	}

	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return "", "", errors.New("invalid session id")
	}

	newAccessToken, err := GenerateAccessToken(
		uint(userIDFloat),
		sessionID,
	)
	if err != nil {
		return "", "", err
	}

	
	newRefreshToken, err := GenerateRefreshToken(
		uint(userIDFloat),
		sessionID,
	)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
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
