package middleware

import (
	"backend_institutions/internal/database"
	"backend_institutions/internal/helper"

	"github.com/gofiber/fiber/v3"
)

func RequirePermission(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {

		userID := c.Locals("user_id")
		if userID == nil {
			return helper.Error(
				c,
				401,
				"Unauthorized",
			)
		}

		var allowed bool

		query := `
			SELECT EXISTS (
				SELECT 1
				FROM users u
				JOIN user_roles ur ON ur.user_id = u.id
				JOIN role_permissions rp ON rp.role_id = ur.role_id
				JOIN permissions p ON p.id = rp.permission_id
				WHERE u.id = ? AND p.name = ?
			)
		`

		err := database.DB.
			Raw(query, userID, permission).
			Scan(&allowed).Error

		if err != nil {
			return helper.Error(
				c,
				500,
				"Failed to check permission",
			)
		}

		if !allowed {
			return helper.Error(
				c,
				403,
				"Forbidden: you do not have permission to perform this action",
			)
		}

		return c.Next()
	}
}
