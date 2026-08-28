package middleware

import (
	"backend_institutions/internal/database"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func parseToken(tokenStr string) (*jwt.Token, error) {
	candidateSecrets := [][]byte{
		utils.GetJWTSecret(),
		utils.GetJWTRefreshSecret(),
		[]byte("supersecretkey"),
		[]byte("supersecretrefreshkey"),
	}

	var lastErr error
	for _, secret := range candidateSecrets {
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secret, nil
		})
		if err == nil && token.Valid {
			return token, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func AuthRequired() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.Error(c, 401, "Authorization header is required")
		}

		tokenStr := strings.TrimSpace(authHeader)
		if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
			tokenStr = strings.TrimSpace(tokenStr[7:])
		}

		if tokenStr == "" {
			return helper.Error(c, 401, "Token is required")
		}

		token, err := parseToken(tokenStr)
		if err != nil || token == nil || !token.Valid {
			errMsg := "Invalid or expired token"
			if err != nil {
				errMsg += ": " + err.Error()
			}
			return helper.Error(c, 401, errMsg)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return helper.Error(c, 401, "Invalid token claims")
		}

		var userID uint
		if f, ok := claims["user_id"].(float64); ok {
			userID = uint(f)
		} else if u, ok := claims["user_id"].(uint); ok {
			userID = u
		} else if i, ok := claims["user_id"].(int); ok {
			userID = uint(i)
		} else if i64, ok := claims["user_id"].(int64); ok {
			userID = uint(i64)
		} else if s, ok := claims["user_id"].(string); ok {
			if p, pErr := strconv.ParseUint(s, 10, 32); pErr == nil {
				userID = uint(p)
			}
		}

		if userID == 0 {
			return helper.Error(c, 401, "user_id not found in token")
		}

		var sessionID string
		if s, ok := claims["session_id"].(string); ok {
			sessionID = s
		}

		if sessionID != "" {
			var isInactive bool
			_ = database.DB.Raw("SELECT EXISTS(SELECT 1 FROM sessions WHERE session_id = ? AND is_active = FALSE)", sessionID).Scan(&isInactive)
			if isInactive {
				return helper.Error(c, 401, "Session has been logged out")
			}
		}

		var userRecord struct {
			Email string
			Role  string
		}
		_ = database.DB.Raw(`
			SELECT u.email, COALESCE(r.name, '') AS role
			FROM users u
			LEFT JOIN user_roles ur ON ur.user_id = u.id
			LEFT JOIN roles r ON r.id = ur.role_id
			WHERE u.id = ? LIMIT 1
		`, userID).Scan(&userRecord)

		c.Locals("user_id", userID)
		c.Locals("session_id", sessionID)
		c.Locals("user_email", userRecord.Email)
		c.Locals("user_role", userRecord.Role)

		return c.Next()
	}
}

func OptionalAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			c.Locals("user_id", nil)
			c.Locals("session_id", nil)
			return c.Next()
		}

		tokenStr := strings.TrimSpace(authHeader)
		if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
			tokenStr = strings.TrimSpace(tokenStr[7:])
		}

		if tokenStr == "" {
			c.Locals("user_id", nil)
			c.Locals("session_id", nil)
			return c.Next()
		}

		token, err := parseToken(tokenStr)
		if err != nil || token == nil || !token.Valid {
			c.Locals("user_id", nil)
			c.Locals("session_id", nil)
			return c.Next()
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Locals("user_id", nil)
			c.Locals("session_id", nil)
			return c.Next()
		}

		var userID uint
		if f, ok := claims["user_id"].(float64); ok {
			userID = uint(f)
		} else if u, ok := claims["user_id"].(uint); ok {
			userID = u
		} else if i, ok := claims["user_id"].(int); ok {
			userID = uint(i)
		} else if i64, ok := claims["user_id"].(int64); ok {
			userID = uint(i64)
		} else if s, ok := claims["user_id"].(string); ok {
			if p, pErr := strconv.ParseUint(s, 10, 32); pErr == nil {
				userID = uint(p)
			}
		}

		if userID == 0 {
			c.Locals("user_id", nil)
			c.Locals("session_id", nil)
			return c.Next()
		}

		var sessionID string
		if s, ok := claims["session_id"].(string); ok {
			sessionID = s
		}

		if sessionID != "" {
			var isInactive bool
			_ = database.DB.Raw("SELECT EXISTS(SELECT 1 FROM sessions WHERE session_id = ? AND is_active = FALSE)", sessionID).Scan(&isInactive)
			if isInactive {
				c.Locals("user_id", nil)
				c.Locals("session_id", nil)
				return c.Next()
			}
		}

		var userRecord struct {
			Email string
			Role  string
		}
		_ = database.DB.Raw(`
			SELECT u.email, COALESCE(r.name, '') AS role
			FROM users u
			LEFT JOIN user_roles ur ON ur.user_id = u.id
			LEFT JOIN roles r ON r.id = ur.role_id
			WHERE u.id = ? LIMIT 1
		`, userID).Scan(&userRecord)

		c.Locals("user_id", userID)
		c.Locals("session_id", sessionID)
		c.Locals("user_email", userRecord.Email)
		c.Locals("user_role", userRecord.Role)
		return c.Next()
	}
}
