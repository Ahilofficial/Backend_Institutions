package middleware

import (
	"backend_institutions/internal/constants"
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

		var isSuperAdmin bool
		_ = database.DB.Raw("SELECT EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = ? AND LOWER(r.name) IN ('super admin', 'super_admin', 'superadmin'))", userID).Scan(&isSuperAdmin)
		if isSuperAdmin {
			return c.Next()
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

		if allowed {
			return c.Next()
		}

		var userRole string
		_ = database.DB.Raw("SELECT LOWER(r.name) FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = ? LIMIT 1", userID).Scan(&userRole)

		switch userRole {
		case "institution admin", "institution_admin", "institution-admin":
			return c.Next()
		case "faculty":
			switch permission {
			case constants.PermissionCreateFaculties, constants.PermissionViewFaculties, constants.PermissionViewIDFaculties, constants.PermissionUpdateFaculties,
				constants.PermissionViewStudents, constants.PermissionViewStudentsID, constants.StudentMonth,
				constants.PermissionViewDepartments, constants.PermissionViewIDDepartments,
				constants.PermissionViewInstitutes, constants.PermissionViewIDInstitutes,
				constants.PermissionViewFees, constants.PermissionViewPayments, constants.PermissionViewIDPayments:
				return c.Next()
			}
		case "student":
			switch permission {
			case constants.PermissionCreateStudents, constants.PermissionViewStudents, constants.PermissionViewStudentsID, constants.PermissionUpdateStudents,
				constants.PermissionViewFaculties, constants.PermissionViewIDFaculties,
				constants.PermissionViewDepartments, constants.PermissionViewIDDepartments,
				constants.PermissionViewInstitutes, constants.PermissionViewIDInstitutes,
				constants.PermissionViewFees, constants.PermissionViewPayments, constants.PermissionViewIDPayments:
				return c.Next()
			}
		case "principal":
			switch permission {
			case constants.PermissionCreatePrincipals, constants.PermissionViewPrincipals, constants.PermissionViewIDPrincipals, constants.PermissionUpdatePrincipals,
				constants.PermissionCreateFaculties, constants.PermissionViewFaculties, constants.PermissionViewIDFaculties, constants.PermissionUpdateFaculties,
				constants.PermissionViewStudents, constants.PermissionViewStudentsID, constants.StudentMonth,
				constants.PermissionViewDepartments, constants.PermissionViewIDDepartments,
				constants.PermissionViewInstitutes, constants.PermissionViewIDInstitutes,
				constants.PermissionViewFees, constants.PermissionViewPayments, constants.PermissionViewIDPayments:
				return c.Next()
			}
		}

		return helper.Error(c, 403, "Access denied")
	}
}
