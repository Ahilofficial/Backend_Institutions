package middleware

import (
	"backend_institutions/internal/database"
	"backend_institutions/internal/helper"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func RequirePermission(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok || userID == 0 {
			return helper.Error(c, 401, "user not authenticated")
		}

		role, _ := c.Locals("user_role").(string)
		if role == "super_admin" || strings.EqualFold(strings.TrimSpace(role), "super_admin") {
			return c.Next()
		}

		var isSuperAdmin bool
		_ = database.DB.Raw(`
			SELECT EXISTS(
				SELECT 1 FROM user_roles ur 
				JOIN roles r ON r.id = ur.role_id 
				WHERE ur.user_id = ? AND LOWER(TRIM(r.name)) = 'super_admin'
			)
		`, userID).Scan(&isSuperAdmin)
		if isSuperAdmin {
			return c.Next()
		}

		var count int64
		_ = database.DB.Raw(`
			SELECT COUNT(*) FROM user_roles ur
			JOIN role_permissions rp ON rp.role_id = ur.role_id
			JOIN permissions p ON p.id = rp.permission_id
			WHERE ur.user_id = ? AND LOWER(TRIM(p.name)) = LOWER(TRIM(?))
		`, userID, permission).Scan(&count)
		if count > 0 {
			return c.Next()
		}

		return helper.Error(c, 403, "Access denied")
	}
}
