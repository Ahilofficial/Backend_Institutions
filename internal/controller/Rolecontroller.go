package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// RoleController handles role, permission, user-role mapping, and role-permission association endpoints
type RoleController struct {
	roleService *services.RoleService
}

// NewRoleController instantiates a new RoleController
func NewRoleController(roleService *services.RoleService) *RoleController {
	return &RoleController{roleService: roleService}
}

// CreateRoleController handles creation of a new role
func (cl *RoleController) CreateRoleController(c fiber.Ctx) error {
	// 1. Bind request body
	var body dto.CreateRoleDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	// 2. Sanitize and validate inputs
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Create role via service
	role, err := cl.roleService.CreateRole(&body)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return success response
	return helper.Success(c, "Role created successfully", role)
}

// GetRoleByIDController fetches role details by ID
func (cl *RoleController) GetRoleByIDController(c fiber.Ctx) error {
	// 1. Parse role ID parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	// 2. Fetch role via service
	role, err := cl.roleService.GetRoleByID(uint(id))
	if err != nil {
		return helper.Error(c, 404, "Role not found")
	}

	// 3. Return response
	return helper.Success(c, "Role retrieved successfully", role)
}

// FetchRoles retrieves paginated list of roles with optional search
func (c *RoleController) FetchRoles(ctx fiber.Ctx) error {
	// 1. Parse pagination parameters
	page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	search := ctx.Query("search")

	// 2. Fetch paginated roles from service
	roles, total, err := c.roleService.FetchRoles(search, page, limit)
	if err != nil {
		return helper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// 3. Return response map
	return helper.Success(ctx, "Roles fetched successfully", fiber.Map{
		"roles": roles,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// FetchPermissions retrieves paginated list of permissions
func (c *RoleController) FetchPermissions(ctx fiber.Ctx) error {
	// 1. Parse pagination parameters
	search := ctx.Query("search")
	page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// 2. Fetch paginated permissions from service
	permissions, total, err := c.roleService.FetchPermissionsService(search, page, limit)
	if err != nil {
		return helper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	// 3. Return response map
	return helper.Success(ctx, "Permissions fetched successfully", fiber.Map{
		"permissions": permissions,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// AssignPermissionsController assigns permission IDs to a role
func (cl *RoleController) AssignPermissionsController(c fiber.Ctx) error {
	// 1. Parse role ID parameter
	idParam := c.Params("id")
	roleID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	// 2. Bind request body
	var body dto.AssignPermissionsDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	// 3. Sanitize and validate
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Assign permissions via service
	if err := cl.roleService.AssignPermissionsToRole(uint(roleID), &body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success
	return helper.Success(c, "Permissions assigned to role successfully", nil)
}

// GetRolePermissionsController retrieves permissions assigned to a role
func (cl *RoleController) GetRolePermissionsController(c fiber.Ctx) error {
	// 1. Parse role ID parameter
	idParam := c.Params("id")
	roleID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	// 2. Fetch role permissions from service
	perms, err := cl.roleService.GetRolePermissions(uint(roleID))
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 3. Return list
	return helper.Success(c, "Role permissions retrieved successfully", perms)
}

// RemovePermissionController detaches a permission from a role
func (cl *RoleController) RemovePermissionController(c fiber.Ctx) error {
	// 1. Parse role ID parameter
	idParam := c.Params("id")
	roleID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	// 2. Parse permission ID parameter
	permParam := c.Params("permissionId")
	permID, err := strconv.ParseUint(permParam, 10, 32)
	if err != nil || permID == 0 {
		return helper.Error(c, 400, "Invalid permission ID")
	}

	// 3. Remove permission via service
	if err := cl.roleService.RemovePermissionFromRole(uint(roleID), uint(permID)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return success
	return helper.Success(c, "Permission removed from role successfully", nil)
}

// UpdateRoleController updates role details (name, description)
func (cl *RoleController) UpdateRoleController(c fiber.Ctx) error {
	// 1. Parse role ID parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	// 2. Bind update body
	var body dto.UpdateRoleDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	// 3. Sanitize and validate
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Update role via service
	if err := cl.roleService.UpdateRole(uint(id), &body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success
	return helper.Success(c, "Role updated successfully", nil)
}

// DeleteRoleController soft deletes a role
func (cl *RoleController) DeleteRoleController(c fiber.Ctx) error {
	// 1. Parse role ID parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid role ID")
	}

	// 2. Delete role via service
	if err := cl.roleService.DeleteRole(uint(id)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Return success
	return helper.Success(c, "Role deleted successfully", nil)
}

// GetPermissionByIDController retrieves permission details by ID
func (cl *RoleController) GetPermissionByIDController(c fiber.Ctx) error {
	// 1. Parse permission ID parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid permission ID")
	}

	// 2. Fetch permission from service
	perm, err := cl.roleService.GetPermissionByID(uint(id))
	if err != nil {
		return helper.Error(c, 404, "Permission not found")
	}

	// 3. Return response
	return helper.Success(c, "Permission retrieved successfully", perm)
}

// DeletePermissionController deletes a permission record
func (cl *RoleController) DeletePermissionController(c fiber.Ctx) error {
	// 1. Parse permission ID parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid permission ID")
	}

	// 2. Delete permission via service
	if err := cl.roleService.DeletePermission(uint(id)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Return success
	return helper.Success(c, "Permission deleted successfully", nil)
}

// FetchUserRolesController retrieves paginated list of user-role mappings
func (cl *RoleController) FetchUserRolesController(c fiber.Ctx) error {
	// 1. Parse pagination parameters
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// 2. Fetch user roles from service
	userRoles, total, err := cl.roleService.FetchUserRoles(page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 3. Return response map
	return helper.Success(c, "User roles fetched successfully", fiber.Map{
		"user_roles": userRoles,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

// CreateUserRoleController associates a user with a role
func (cl *RoleController) CreateUserRoleController(c fiber.Ctx) error {
	// 1. Bind request body
	var body dto.CreateUserRoleDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	// 2. Sanitize and validate
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Create mapping via service
	if err := cl.roleService.CreateUserRole(&body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return success
	return helper.Success(c, "User role created successfully", nil)
}

// GetUserRoleByIDController retrieves specific user-role mapping
func (cl *RoleController) GetUserRoleByIDController(c fiber.Ctx) error {
	// 1. Parse user and role ID parameters
	userIDParam := c.Params("userId")
	userID, err1 := strconv.ParseUint(userIDParam, 10, 32)
	roleIDParam := c.Params("roleId")
	roleID, err2 := strconv.ParseUint(roleIDParam, 10, 32)

	if err1 != nil || userID == 0 || err2 != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid user ID or role ID")
	}

	// 2. Fetch user-role mapping from service
	userRole, err := cl.roleService.GetUserRoleByID(uint(userID), uint(roleID))
	if err != nil {
		return helper.Error(c, 404, "User role mapping not found")
	}

	// 3. Return response
	return helper.Success(c, "User role mapping retrieved successfully", userRole)
}

// UpdateUserRoleController updates a user-role mapping
func (cl *RoleController) UpdateUserRoleController(c fiber.Ctx) error {
	// 1. Parse user and role ID parameters
	userIDParam := c.Params("userId")
	userID, err1 := strconv.ParseUint(userIDParam, 10, 32)
	roleIDParam := c.Params("roleId")
	roleID, err2 := strconv.ParseUint(roleIDParam, 10, 32)

	if err1 != nil || userID == 0 || err2 != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid user ID or role ID")
	}

	// 2. Bind update body
	var body dto.UpdateUserRoleDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	// 3. Sanitize and validate
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Update mapping via service
	if err := cl.roleService.UpdateUserRole(uint(userID), uint(roleID), &body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success
	return helper.Success(c, "User role mapping updated successfully", nil)
}

// DeleteUserRoleController removes user-role mapping
func (cl *RoleController) DeleteUserRoleController(c fiber.Ctx) error {
	// 1. Parse user and role ID parameters
	userIDParam := c.Params("userId")
	userID, err1 := strconv.ParseUint(userIDParam, 10, 32)
	roleIDParam := c.Params("roleId")
	roleID, err2 := strconv.ParseUint(roleIDParam, 10, 32)

	if err1 != nil || userID == 0 || err2 != nil || roleID == 0 {
		return helper.Error(c, 400, "Invalid user ID or role ID")
	}

	// 2. Delete user-role mapping via service
	if err := cl.roleService.DeleteUserRole(uint(userID), uint(roleID)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Return success
	return helper.Success(c, "User role mapping deleted successfully", nil)
}

// FetchRolePermissionsController retrieves paginated list of role-permission associations
func (cl *RoleController) FetchRolePermissionsController(c fiber.Ctx) error {
	// 1. Parse pagination parameters
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// 2. Fetch role permissions list from service
	rolePerms, total, err := cl.roleService.FetchRolePermissions(page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 3. Return response map
	return helper.Success(c, "Role permissions fetched successfully", fiber.Map{
		"role_permissions": rolePerms,
		"total":            total,
		"page":             page,
		"limit":            limit,
	})
}

// CreateRolePermissionController creates a role-permission association
func (cl *RoleController) CreateRolePermissionController(c fiber.Ctx) error {
	// 1. Bind request body
	var body dto.CreateRolePermissionDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	// 2. Sanitize and validate
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Create role permission association via service
	if err := cl.roleService.CreateRolePermission(&body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return success
	return helper.Success(c, "Role permission created successfully", nil)
}

// GetRolePermissionByIDController retrieves role permission mapping details
func (cl *RoleController) GetRolePermissionByIDController(c fiber.Ctx) error {
	// 1. Parse role and permission ID parameters
	roleIDParam := c.Params("roleId")
	roleID, err1 := strconv.ParseUint(roleIDParam, 10, 32)
	permIDParam := c.Params("permissionId")
	permID, err2 := strconv.ParseUint(permIDParam, 10, 32)

	if err1 != nil || roleID == 0 || err2 != nil || permID == 0 {
		return helper.Error(c, 400, "Invalid role ID or permission ID")
	}

	// 2. Fetch mapping via service
	rolePerm, err := cl.roleService.GetRolePermissionByID(uint(roleID), uint(permID))
	if err != nil {
		return helper.Error(c, 404, "Role permission mapping not found")
	}

	// 3. Return response
	return helper.Success(c, "Role permission mapping retrieved successfully", rolePerm)
}

// UpdateRolePermissionController updates a role-permission association
func (cl *RoleController) UpdateRolePermissionController(c fiber.Ctx) error {
	// 1. Parse role and permission ID parameters
	roleIDParam := c.Params("roleId")
	roleID, err1 := strconv.ParseUint(roleIDParam, 10, 32)
	permIDParam := c.Params("permissionId")
	permID, err2 := strconv.ParseUint(permIDParam, 10, 32)

	if err1 != nil || roleID == 0 || err2 != nil || permID == 0 {
		return helper.Error(c, 400, "Invalid role ID or permission ID")
	}

	// 2. Bind update body
	var body dto.UpdateRolePermissionDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	// 3. Sanitize and validate
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Update mapping via service
	if err := cl.roleService.UpdateRolePermission(uint(roleID), uint(permID), &body); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success
	return helper.Success(c, "Role permission mapping updated successfully", nil)
}

// DeleteRolePermissionController deletes a role permission mapping
func (cl *RoleController) DeleteRolePermissionController(c fiber.Ctx) error {
	// 1. Parse role and permission ID parameters
	roleIDParam := c.Params("roleId")
	roleID, err1 := strconv.ParseUint(roleIDParam, 10, 32)
	permIDParam := c.Params("permissionId")
	permID, err2 := strconv.ParseUint(permIDParam, 10, 32)

	if err1 != nil || roleID == 0 || err2 != nil || permID == 0 {
		return helper.Error(c, 400, "Invalid role ID or permission ID")
	}

	// 2. Delete mapping via service
	if err := cl.roleService.DeleteRolePermission(uint(roleID), uint(permID)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Return success
	return helper.Success(c, "Role permission mapping deleted successfully", nil)
}

// GetUserRolesByUserIDController retrieves all roles assigned to a user
func (cl *RoleController) GetUserRolesByUserIDController(c fiber.Ctx) error {
	// 1. Parse user ID parameter
	userIDParam := c.Params("userId")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil || userID == 0 {
		return helper.Error(c, 400, "Invalid user ID")
	}

	// 2. Fetch user roles from service
	userRoles, err := cl.roleService.GetUserRolesByUserID(uint(userID))
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 3. Return response
	return helper.Success(c, "User roles retrieved successfully", userRoles)
}

// FetchAllRoles fetches all roles with their associated permissions
func (cl *RoleController) FetchAllRoles(c fiber.Ctx) error {
	// 1. Fetch roles and permissions from service
	roles, err := cl.roleService.FetchAllRolesPermissions()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 2. Return response
	return helper.Success(c, "Roles fetched successfully", roles)
}
