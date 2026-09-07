package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
)

type RoleService struct {
	rolerepo *repository.RoleRepository
}

func NewRoleService(rolerepo *repository.RoleRepository) *RoleService {
	return &RoleService{rolerepo: rolerepo}
}

func (s *RoleService) CreateRole(createDTO *dto.CreateRoleDTO) (dto.RoleResponseDTO, error) {

	role := model.Role{
		Name: createDTO.Name,
	}

	err := s.rolerepo.CreateRole(&role)
	if err != nil {
		return dto.RoleResponseDTO{}, err
	}

	return dto.ToRoleResponseDTO(&role), nil
}

func (s *RoleService) FetchRoles(search string, page, limit int) ([]model.Role, int64, error) {
	return s.rolerepo.FetchRoles(search, page, limit)
}

func (s *RoleService) FetchPermissionsService(search string, page, limit int) ([]model.Permission, int64, error) {
	return s.rolerepo.Permissions(search, page, limit)
}

func (s *RoleService) GetRoleByID(id uint) (dto.RoleResponseDTO, error) {

	role, err := s.rolerepo.GetRoleByID(id)
	if err != nil {
		return dto.RoleResponseDTO{}, err
	}

	return dto.ToRoleResponseDTO(&role), nil
}

func (s *RoleService) AssignPermissionsToRole(roleID uint, assignDTO *dto.AssignPermissionsDTO) error {
	return s.rolerepo.AssignPermissionsToRole(roleID, assignDTO.PermissionIDs, assignDTO.PermissionNames)
}

func (s *RoleService) GetRolePermissions(roleID uint) ([]dto.PermissionResponseDTO, error) {

	perms, err := s.rolerepo.GetRolePermissions(roleID)
	if err != nil {
		return nil, err
	}

	return dto.ToPermissionResponseListDTO(perms), nil
}

func (s *RoleService) RemovePermissionFromRole(roleID uint, permissionID uint) error {
	return s.rolerepo.RemovePermissionFromRole(roleID, permissionID)
}

func (s *RoleService) UpdateRole(id uint, updateDTO *dto.UpdateRoleDTO) error {
	return s.rolerepo.UpdateRole(id, updateDTO.Name)
}

func (s *RoleService) DeleteRole(id uint) error {
	return s.rolerepo.DeleteRole(id)
}

func (s *RoleService) GetPermissionByID(id uint) (dto.PermissionResponseDTO, error) {

	perm, err := s.rolerepo.GetPermissionByID(id)
	if err != nil {
		return dto.PermissionResponseDTO{}, err
	}

	return dto.ToPermissionResponseDTO(&perm), nil
}

func (s *RoleService) DeletePermission(id uint) error {
	return s.rolerepo.DeletePermission(id)
}

func (s *RoleService) FetchUserRoles(page, limit int) ([]map[string]any, int64, error) {
	return s.rolerepo.FetchUserRoles(page, limit)
}

func (s *RoleService) CreateUserRole(dto *dto.CreateUserRoleDTO) error {
	return s.rolerepo.CreateUserRole(dto.UserID, dto.RoleID)
}

func (s *RoleService) GetUserRoleByID(userID, roleID uint) (map[string]any, error) {
	return s.rolerepo.GetUserRoleByID(userID, roleID)
}

func (s *RoleService) UpdateUserRole(userID, roleID uint, dto *dto.UpdateUserRoleDTO) error {
	return s.rolerepo.UpdateUserRole(userID, roleID, dto.RoleID)
}

func (s *RoleService) DeleteUserRole(userID, roleID uint) error {
	return s.rolerepo.DeleteUserRole(userID, roleID)
}

func (s *RoleService) FetchRolePermissions(page, limit int) ([]map[string]any, int64, error) {
	return s.rolerepo.FetchRolePermissions(page, limit)
}

func (s *RoleService) CreateRolePermission(dto *dto.CreateRolePermissionDTO) error {
	return s.rolerepo.CreateRolePermission(dto.RoleID, dto.PermissionID)
}

func (s *RoleService) GetRolePermissionByID(roleID, permissionID uint) (map[string]any, error) {
	return s.rolerepo.GetRolePermissionByID(roleID, permissionID)
}

func (s *RoleService) UpdateRolePermission(roleID, permissionID uint, dto *dto.UpdateRolePermissionDTO) error {
	return s.rolerepo.UpdateRolePermission(roleID, permissionID, dto.PermissionID)
}

func (s *RoleService) DeleteRolePermission(roleID, permissionID uint) error {
	return s.rolerepo.DeleteRolePermission(roleID, permissionID)
}

func (s *RoleService) GetUserRolesByUserID(userID uint) (*model.User, error) {
	return s.rolerepo.GetUserRolesByUserID(userID)
}

func (s *RoleService) FetchAllRolesPermissions() ([]dto.RolesDTOResponse, error) {
	return s.rolerepo.FetchAllRolesPermissions()
}
