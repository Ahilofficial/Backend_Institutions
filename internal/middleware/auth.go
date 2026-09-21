package middleware

import (
	"backend_institutions/internal/helper"
	"backend_institutions/internal/utils"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() fiber.Handler {
	return func(c fiber.Ctx) error {

		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return helper.Error(c, 401, "Authorization header is required")
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenStr == "" {
			return helper.Error(c, 401, "Token is required")
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}

			return utils.GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			return helper.Error(c, 401, "Invalid or expired token")
		}

		claims,ok:= token.Claims.(jwt.MapClaims)
		if !ok {
			return helper.Error(c, 401, "Invalid token claims")
		}
		fmt.Println(claims)
		

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return helper.Error(c, 401, "user_id not found in token")
		}

		userID := uint(userIDFloat)

		if userID == 0 {
			return helper.Error(c, 401, "Invalid user ID")
		}

		c.Locals("user_id", userID)

		return c.Next()
		
	}
	
}