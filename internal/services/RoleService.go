package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
)

// RoleService provides business logic for roles, permissions, and role mappings
type RoleService struct {
	rolerepo *repository.RoleRepository
}

// NewRoleService initializes a new instance of RoleService
func NewRoleService(rolerepo *repository.RoleRepository) *RoleService {
	return &RoleService{rolerepo: rolerepo}
}

// CreateRole creates a new role entity
func (s *RoleService) CreateRole(createDTO *dto.CreateRoleDTO) (dto.RoleResponseDTO, error) {
	// 1. Initialize role model
	role := model.Role{
		Name: createDTO.Name,
	}

	// 2. Persist role in repository
	err := s.rolerepo.CreateRole(&role)
	if err != nil {
		return dto.RoleResponseDTO{}, err
	}

	// 3. Return response DTO
	return dto.ToRoleResponseDTO(&role), nil
}

// FetchRoles retrieves paginated roles matching optional search query
func (s *RoleService) FetchRoles(search string, page, limit int) ([]model.Role, int64, error) {
	return s.rolerepo.FetchRoles(search, page, limit)
}

// FetchPermissionsService retrieves paginated system permissions
func (s *RoleService) FetchPermissionsService(search string, page, limit int) ([]model.Permission, int64, error) {
	return s.rolerepo.Permissions(search, page, limit)
}

// GetRoleByID fetches a single role by its primary key ID
func (s *RoleService) GetRoleByID(id uint) (dto.RoleResponseDTO, error) {
	// 1. Query role by ID
	role, err := s.rolerepo.GetRoleByID(id)
	if err != nil {
		return dto.RoleResponseDTO{}, err
	}

	// 2. Convert and return DTO
	return dto.ToRoleResponseDTO(&role), nil
}

// AssignPermissionsToRole links permissions to a role by IDs or names
func (s *RoleService) AssignPermissionsToRole(roleID uint, assignDTO *dto.AssignPermissionsDTO) error {
	return s.rolerepo.AssignPermissionsToRole(roleID, assignDTO.PermissionIDs, assignDTO.PermissionNames)
}

// GetRolePermissions fetches all permissions assigned to a role
func (s *RoleService) GetRolePermissions(roleID uint) ([]dto.PermissionResponseDTO, error) {
	// 1. Fetch permissions from repository
	perms, err := s.rolerepo.GetRolePermissions(roleID)
	if err != nil {
		return nil, err
	}

	// 2. Convert and return DTO slice
	return dto.ToPermissionResponseListDTO(perms), nil
}

// RemovePermissionFromRole detaches a permission from a role
func (s *RoleService) RemovePermissionFromRole(roleID uint, permissionID uint) error {
	return s.rolerepo.RemovePermissionFromRole(roleID, permissionID)
}

// UpdateRole updates role name
func (s *RoleService) UpdateRole(id uint, updateDTO *dto.UpdateRoleDTO) error {
	return s.rolerepo.UpdateRole(id, updateDTO.Name)
}

// DeleteRole soft deletes a role
func (s *RoleService) DeleteRole(id uint) error {
	return s.rolerepo.DeleteRole(id)
}

// GetPermissionByID retrieves permission details by ID
func (s *RoleService) GetPermissionByID(id uint) (dto.PermissionResponseDTO, error) {
	// 1. Fetch permission by ID
	perm, err := s.rolerepo.GetPermissionByID(id)
	if err != nil {
		return dto.PermissionResponseDTO{}, err
	}

	// 2. Return response DTO
	return dto.ToPermissionResponseDTO(&perm), nil
}

// DeletePermission deletes a permission record
func (s *RoleService) DeletePermission(id uint) error {
	return s.rolerepo.DeletePermission(id)
}

// FetchUserRoles retrieves paginated user-role associations
func (s *RoleService) FetchUserRoles(page, limit int) ([]map[string]any, int64, error) {
	return s.rolerepo.FetchUserRoles(page, limit)
}

// CreateUserRole creates a mapping between user and role
func (s *RoleService) CreateUserRole(dto *dto.CreateUserRoleDTO) error {
	return s.rolerepo.CreateUserRole(dto.UserID, dto.RoleID)
}

// GetUserRoleByID fetches a user-role mapping by compound keys
func (s *RoleService) GetUserRoleByID(userID, roleID uint) (map[string]any, error) {
	return s.rolerepo.GetUserRoleByID(userID, roleID)
}

// UpdateUserRole updates the role assigned in a user-role mapping
func (s *RoleService) UpdateUserRole(userID, roleID uint, dto *dto.UpdateUserRoleDTO) error {
	return s.rolerepo.UpdateUserRole(userID, roleID, dto.RoleID)
}

// DeleteUserRole deletes a user-role mapping
func (s *RoleService) DeleteUserRole(userID, roleID uint) error {
	return s.rolerepo.DeleteUserRole(userID, roleID)
}

// FetchRolePermissions retrieves paginated role-permission mappings
func (s *RoleService) FetchRolePermissions(page, limit int) ([]map[string]any, int64, error) {
	return s.rolerepo.FetchRolePermissions(page, limit)
}

// CreateRolePermission creates a role-permission mapping
func (s *RoleService) CreateRolePermission(dto *dto.CreateRolePermissionDTO) error {
	return s.rolerepo.CreateRolePermission(dto.RoleID, dto.PermissionID)
}

// GetRolePermissionByID fetches a role-permission mapping by compound keys
func (s *RoleService) GetRolePermissionByID(roleID, permissionID uint) (map[string]any, error) {
	return s.rolerepo.GetRolePermissionByID(roleID, permissionID)
}

// UpdateRolePermission updates permission assigned in a role-permission mapping
func (s *RoleService) UpdateRolePermission(roleID, permissionID uint, dto *dto.UpdateRolePermissionDTO) error {
	return s.rolerepo.UpdateRolePermission(roleID, permissionID, dto.PermissionID)
}

// DeleteRolePermission deletes a role-permission mapping
func (s *RoleService) DeleteRolePermission(roleID, permissionID uint) error {
	return s.rolerepo.DeleteRolePermission(roleID, permissionID)
}

// GetUserRolesByUserID fetches user and preloads their assigned roles
func (s *RoleService) GetUserRolesByUserID(userID uint) (*model.User, error) {
	return s.rolerepo.GetUserRolesByUserID(userID)
}

// FetchAllRolesPermissions retrieves all roles with associated permissions
func (s *RoleService) FetchAllRolesPermissions() ([]dto.RolesDTOResponse, error) {
	return s.rolerepo.FetchAllRolesPermissions()
}
