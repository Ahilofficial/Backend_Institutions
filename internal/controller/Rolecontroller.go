package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type RoleController struct {
	roleService *services.RoleService
}

func NewRoleController(roleService *services.RoleService) *RoleController {
	return &RoleController{roleService: roleService}
}

func (cl *RoleController) CreateRoleController(c fiber.Ctx) error {
	var body dto.CreateRoleDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	role, err := cl.roleService.CreateRole(&body)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Role created successfully", role)
}

func (cl *RoleController) GetRoleByIDController(c fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	role, err := cl.roleService.GetRoleByID(uint(id))
	if err != nil {
		return helper.Error(c, 404, "Role not found")
	}

	return helper.Success(c, "Role retrieved successfully", role)
}

func (c *RoleController) FetchRoles(ctx fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	search := ctx.Query("search")

	roles, total, err := c.roleService.FetchRoles(search, page, limit)
	if err != nil {
		return helper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return helper.Success(ctx, "Roles fetched successfully", fiber.Map{
		"roles": roles,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (c *RoleController) FetchPermissions(ctx fiber.Ctx) error {

	search := ctx.Query("search")
	page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	permissions, total, err := c.roleService.FetchPermissionsService(search, page, limit)
	if err != nil {
		return helper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return helper.Success(ctx, "Permissions fetched successfully", fiber.Map{
		"permissions": permissions,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

func (cl *RoleController) AssignPermissionsController(c fiber.Ctx) error {
	idParam := c.Params("id")
	roleID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	var body dto.AssignPermissionsDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.roleService.AssignPermissionsToRole(uint(roleID), &body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Permissions assigned to role successfully", nil)
}

func (cl *RoleController) GetRolePermissionsController(c fiber.Ctx) error {
	idParam := c.Params("id")
	roleID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	perms, err := cl.roleService.GetRolePermissions(uint(roleID))
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, "Role permissions retrieved successfully", perms)
}

func (cl *RoleController) RemovePermissionController(c fiber.Ctx) error {
	idParam := c.Params("id")
	roleID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	permParam := c.Params("permissionId")
	permID, err := strconv.ParseUint(permParam, 10, 32)
	if err != nil || permID == 0 {
		return helper.Error(c, 400, "Invalid permission ID")
	}

	if err := cl.roleService.RemovePermissionFromRole(uint(roleID), uint(permID)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Permission removed from role successfully", nil)
}

func (cl *RoleController) UpdateRoleController(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	var body dto.UpdateRoleDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.roleService.UpdateRole(uint(id), &body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Role updated successfully", nil)
}

func (cl *RoleController) DeleteRoleController(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	if err := cl.roleService.DeleteRole(uint(id)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Role deleted successfully", nil)
}

func (cl *RoleController) GetPermissionByIDController(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid permission ID")
	}

	perm, err := cl.roleService.GetPermissionByID(uint(id))
	if err != nil {
		return helper.Error(c, 404, "Permission not found")
	}

	return helper.Success(c, "Permission retrieved successfully", perm)
}

func (cl *RoleController) DeletePermissionController(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid permission ID")
	}

	if err := cl.roleService.DeletePermission(uint(id)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Permission deleted successfully", nil)
}

func (cl *RoleController) FetchUserRolesController(c fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	userRoles, total, err := cl.roleService.FetchUserRoles(page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, "User roles fetched successfully", fiber.Map{
		"user_roles": userRoles,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

func (cl *RoleController) CreateUserRoleController(c fiber.Ctx) error {
	var body dto.CreateUserRoleDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.roleService.CreateUserRole(&body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "User role created successfully", nil)
}

func (cl *RoleController) GetUserRoleByIDController(c fiber.Ctx) error {
	userIDParam := c.Params("userId")
	userID, err1 := strconv.ParseUint(userIDParam, 10, 32)
	roleIDParam := c.Params("roleId")
	roleID, err2 := strconv.ParseUint(roleIDParam, 10, 32)

	if err1 != nil || userID == 0 || err2 != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid user ID or role ID")
	}

	userRole, err := cl.roleService.GetUserRoleByID(uint(userID), uint(roleID))
	if err != nil {
		return helper.Error(c, 404, "User role mapping not found")
	}

	return helper.Success(c, "User role mapping retrieved successfully", userRole)
}

func (cl *RoleController) UpdateUserRoleController(c fiber.Ctx) error {
	userIDParam := c.Params("userId")
	userID, err1 := strconv.ParseUint(userIDParam, 10, 32)
	roleIDParam := c.Params("roleId")
	roleID, err2 := strconv.ParseUint(roleIDParam, 10, 32)

	if err1 != nil || userID == 0 || err2 != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid user ID or role ID")
	}

	var body dto.UpdateUserRoleDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.roleService.UpdateUserRole(uint(userID), uint(roleID), &body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "User role mapping updated successfully", nil)
}

func (cl *RoleController) DeleteUserRoleController(c fiber.Ctx) error {
	userIDParam := c.Params("userId")
	userID, err1 := strconv.ParseUint(userIDParam, 10, 32)
	roleIDParam := c.Params("roleId")
	roleID, err2 := strconv.ParseUint(roleIDParam, 10, 32)

	if err1 != nil || userID == 0 || err2 != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid user ID or role ID")
	}

	if err := cl.roleService.DeleteUserRole(uint(userID), uint(roleID)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "User role mapping deleted successfully", nil)
}

func (cl *RoleController) FetchRolePermissionsController(c fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	rolePerms, total, err := cl.roleService.FetchRolePermissions(page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, "Role permissions fetched successfully", fiber.Map{
		"role_permissions": rolePerms,
		"total":            total,
		"page":             page,
		"limit":            limit,
	})
}

func (cl *RoleController) CreateRolePermissionController(c fiber.Ctx) error {
	var body dto.CreateRolePermissionDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.roleService.CreateRolePermission(&body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Role permission created successfully", nil)
}

func (cl *RoleController) GetRolePermissionByIDController(c fiber.Ctx) error {
	roleIDParam := c.Params("roleId")
	roleID, err1 := strconv.ParseUint(roleIDParam, 10, 32)
	permIDParam := c.Params("permissionId")
	permID, err2 := strconv.ParseUint(permIDParam, 10, 32)

	if err1 != nil || roleID == 0 || err2 != nil || permID == 0 {
		return helper.Error(c, 400, "Invalid role ID or permission ID")
	}

	rolePerm, err := cl.roleService.GetRolePermissionByID(uint(roleID), uint(permID))
	if err != nil {
		return helper.Error(c, 404, "Role permission mapping not found")
	}

	return helper.Success(c, "Role permission mapping retrieved successfully", rolePerm)
}

func (cl *RoleController) UpdateRolePermissionController(c fiber.Ctx) error {
	roleIDParam := c.Params("roleId")
	roleID, err1 := strconv.ParseUint(roleIDParam, 10, 32)
	permIDParam := c.Params("permissionId")
	permID, err2 := strconv.ParseUint(permIDParam, 10, 32)

	if err1 != nil || roleID == 0 || err2 != nil || permID == 0 {
		return helper.Error(c, 400, "Invalid role ID or permission ID")
	}

	var body dto.UpdateRolePermissionDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.roleService.UpdateRolePermission(uint(roleID), uint(permID), &body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Role permission mapping updated successfully", nil)
}

func (cl *RoleController) DeleteRolePermissionController(c fiber.Ctx) error {
	roleIDParam := c.Params("roleId")
	roleID, err1 := strconv.ParseUint(roleIDParam, 10, 32)
	permIDParam := c.Params("permissionId")
	permID, err2 := strconv.ParseUint(permIDParam, 10, 32)

	if err1 != nil || roleID == 0 || err2 != nil || permID == 0 {
		return helper.Error(c, 400, "Invalid role ID or permission ID")
	}

	if err := cl.roleService.DeleteRolePermission(uint(roleID), uint(permID)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Role permission mapping deleted successfully", nil)
}

func (cl *RoleController) GetUserRolesByUserIDController(c fiber.Ctx) error {
	userIDParam := c.Params("userId")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil || userID == 0 {
		return helper.Error(c, 400, "Invalid user ID")
	}

	userRoles, err := cl.roleService.GetUserRolesByUserID(uint(userID))
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, "User roles retrieved successfully", userRoles)
}

func (cl *RoleController) FetchAllRoles(c fiber.Ctx) error {

	roles, err := cl.roleService.FetchAllRolesPermissions()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, "Roles fetched successfully", roles)
}
