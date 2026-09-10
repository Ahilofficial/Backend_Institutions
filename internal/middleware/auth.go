package middleware

import (
	"backend_institutions/internal/helper"
	"backend_institutions/internal/utils"
	// "strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() fiber.Handler {
	return func(c fiber.Ctx) error {

		// Get Authorization header
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return helper.Error(c, 401, "Authorization header is required")
		}

		// Remove "Bearer "
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenStr == "" {
			return helper.Error(c, 401, "Token is required")
		}

		// Parse token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {

			// Make sure algorithm is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}

			return utils.GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			return helper.Error(c, 401, "Invalid or expired token")
		}

		// Get claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return helper.Error(c, 401, "Invalid token claims")
		}

		// Get user_id
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return helper.Error(c, 401, "user_id not found in token")
		}

		userID := uint(userIDFloat)

		if userID == 0 {
			return helper.Error(c, 401, "Invalid user ID")
		}

		// Store user ID for controllers
		c.Locals("user_id", userID)

		return c.Next()
	}
}