package repository

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// RoleRepository manages database persistence and queries for roles, permissions, and mappings
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository instantiates a new RoleRepository
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// CreateRole inserts a new role record
func (r *RoleRepository) CreateRole(role *model.Role) error {
	return r.db.Exec(
		"INSERT INTO roles (name) VALUES (?)",
		role.Name,
	).Error
}

// FetchRoles retrieves paginated roles matching optional search query
func (r *RoleRepository) FetchRoles(search string, page, limit int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	offset := (page - 1) * limit
	searchPattern := "%" + search + "%"

	// 1. Count matching roles
	err := r.db.Raw(`
		SELECT COUNT(*)
		FROM roles
		WHERE deleted_at IS NULL
		AND (? = '' OR name LIKE ?)
	`, search, searchPattern).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 2. Fetch page records
	err = r.db.Raw(`
		SELECT *
		FROM roles
		WHERE deleted_at IS NULL
		AND (? = '' OR name LIKE ?)
		ORDER BY id
		LIMIT ? OFFSET ?
	`, search, searchPattern, limit, offset).Scan(&roles).Error
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// GetRoleByID retrieves a role record by primary key ID
func (r *RoleRepository) GetRoleByID(id uint) (model.Role, error) {
	var role model.Role
	err := r.db.Raw("SELECT id, name, created_at, updated_at FROM roles WHERE id = ? LIMIT 1", id).Scan(&role).Error
	if err != nil {
		return role, err
	}
	if role.ID == 0 {
		return role, gorm.ErrRecordNotFound
	}
	return role, nil
}

// AssignPermissionsToRole links permission records to a role ID
func (r *RoleRepository) AssignPermissionsToRole(roleID uint, permissionIDs []uint, permissionNames []string) error {
	// 1. Verify role existence
	role, err := r.GetRoleByID(roleID)
	if err != nil || role.ID == 0 {
		return errors.New("role not found")
	}

	var targetIDs []uint
	if len(permissionIDs) > 0 {
		targetIDs = append(targetIDs, permissionIDs...)
	}

	// 2. Look up IDs for named permissions
	if len(permissionNames) > 0 {
		var nameIDs []uint
		err := r.db.Raw("SELECT id FROM permissions WHERE name IN ?", permissionNames).Scan(&nameIDs).Error
		if err == nil {
			targetIDs = append(targetIDs, nameIDs...)
		}
	}

	if len(targetIDs) == 0 {
		return errors.New("no valid permissions found to assign")
	}

	// 3. Insert role_permissions association records
	for _, pid := range targetIDs {
		_ = r.db.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, pid)
	}

	return nil
}

// GetRolePermissions fetches all permissions assigned to a role
func (r *RoleRepository) GetRolePermissions(roleID uint) ([]model.Permission, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("id = ? AND deleted_at IS NULL", roleID).First(&role).Error
	if err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

// Permissions retrieves paginated system permissions
func (r *RoleRepository) Permissions(search string, page, limit int) ([]model.Permission, int64, error) {
	var permissions []model.Permission
	var total int64

	offset := (page - 1) * limit
	searchPattern := "%" + search + "%"

	// 1. Count matching permissions
	err := r.db.Raw(`
		SELECT COUNT(*)
		FROM permissions
		WHERE deleted_at IS NULL
		AND (? = '' OR name LIKE ?)
	`, search, searchPattern).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 2. Query paginated permissions
	err = r.db.Raw(`
		SELECT id, name
		FROM permissions
		WHERE deleted_at IS NULL
		AND (? = '' OR name LIKE ?)
		ORDER BY id
		LIMIT ? OFFSET ?
	`, search, searchPattern, limit, offset).Scan(&permissions).Error

	if err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// RemovePermissionFromRole removes a permission link from a role
func (r *RoleRepository) RemovePermissionFromRole(roleID uint, permissionID uint) error {
	res := r.db.Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("role permission mapping not found")
	}
	return nil
}

// UpdateRole updates role name
func (r *RoleRepository) UpdateRole(id uint, name string) error {
	return r.db.Exec("UPDATE roles SET name = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", name, id).Error
}

// DeleteRole soft deletes a role
func (r *RoleRepository) DeleteRole(id uint) error {
	return r.db.Exec("UPDATE roles SET deleted_at = NOW() WHERE id = ?", id).Error
}

// GetPermissionByID retrieves permission details by ID
func (r *RoleRepository) GetPermissionByID(id uint) (model.Permission, error) {
	var perm model.Permission
	err := r.db.Raw("SELECT id, name, created_at, updated_at FROM permissions WHERE id = ? AND deleted_at IS NULL LIMIT 1", id).Scan(&perm).Error
	if err != nil {
		return perm, err
	}
	if perm.ID == 0 {
		return perm, gorm.ErrRecordNotFound
	}
	return perm, nil
}

// DeletePermission soft deletes a permission
func (r *RoleRepository) DeletePermission(id uint) error {
	return r.db.Exec("UPDATE permissions SET deleted_at = NOW() WHERE id = ?", id).Error
}

// FetchUserRoles retrieves paginated user-role associations
func (r *RoleRepository) FetchUserRoles(page, limit int) ([]map[string]any, int64, error) {
	var results []map[string]any
	var total int64

	// 1. Count total user_roles
	err := r.db.Table("user_roles").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	// 2. Query paginated user_roles
	err = r.db.Raw(`
		SELECT ur.user_id, ur.role_id 
		FROM user_roles ur
		LIMIT ? OFFSET ?
	`, limit, offset).Scan(&results).Error

	return results, total, err
}

// CreateUserRole creates a mapping between user and role
func (r *RoleRepository) CreateUserRole(userID, roleID uint) error {
	var userCount, roleCount int64
	_ = r.db.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Count(&userCount)
	_ = r.db.Table("roles").Where("id = ? AND deleted_at IS NULL", roleID).Count(&roleCount)
	if userCount == 0 {
		return errors.New("user not found")
	}
	if roleCount == 0 {
		return errors.New("role not found")
	}

	return r.db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

// GetUserRoleByID retrieves user-role mapping by compound keys
func (r *RoleRepository) GetUserRoleByID(userID, roleID uint) (map[string]any, error) {
	var count int64
	r.db.Table("user_roles").Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count)
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var result map[string]any
	err := r.db.Raw(`
		SELECT ur.user_id, ur.role_id 
		FROM user_roles ur
		WHERE ur.user_id = ? AND ur.role_id = ?
		LIMIT 1
	`, userID, roleID).Scan(&result).Error
	return result, err
}

// UpdateUserRole updates assigned role in user-role mapping
func (r *RoleRepository) UpdateUserRole(userID, roleID, newRoleID uint) error {
	var roleCount int64
	_ = r.db.Table("roles").Where("id = ? AND deleted_at IS NULL", newRoleID).Count(&roleCount)
	if roleCount == 0 {
		return errors.New("new role not found")
	}

	return r.db.Exec("UPDATE user_roles SET role_id = ? WHERE user_id = ? AND role_id = ?", newRoleID, userID, roleID).Error
}

// DeleteUserRole removes user-role mapping
func (r *RoleRepository) DeleteUserRole(userID, roleID uint) error {
	res := r.db.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user role mapping not found")
	}
	return nil
}

// FetchRolePermissions retrieves paginated role-permission mappings
func (r *RoleRepository) FetchRolePermissions(page, limit int) ([]map[string]any, int64, error) {
	var results []map[string]any
	var total int64

	// 1. Count mappings
	err := r.db.Table("role_permissions").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	// 2. Query page
	err = r.db.Raw(`
		SELECT rp.role_id, rp.permission_id 
		FROM role_permissions rp
		LIMIT ? OFFSET ?
	`, limit, offset).Scan(&results).Error

	return results, total, err
}

// CreateRolePermission associates a permission with a role
func (r *RoleRepository) CreateRolePermission(roleID, permissionID uint) error {
	var roleCount, permCount int64
	_ = r.db.Table("roles").Where("id = ? AND deleted_at IS NULL", roleID).Count(&roleCount)
	_ = r.db.Table("permissions").Where("id = ? AND deleted_at IS NULL", permissionID).Count(&permCount)
	if roleCount == 0 {
		return errors.New("role not found")
	}
	if permCount == 0 {
		return errors.New("permission not found")
	}

	return r.db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).Error
}

// GetRolePermissionByID retrieves a role permission mapping by compound keys
func (r *RoleRepository) GetRolePermissionByID(roleID, permissionID uint) (map[string]any, error) {
	var count int64
	r.db.Table("role_permissions").Where("role_id = ? AND permission_id = ?", roleID, permissionID).Count(&count)
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var result map[string]any
	err := r.db.Raw(`
		SELECT rp.role_id, rp.permission_id 
		FROM role_permissions rp
		WHERE rp.role_id = ? AND rp.permission_id = ?
		LIMIT 1
	`, roleID, permissionID).Scan(&result).Error
	return result, err
}

// UpdateRolePermission updates permission in a role permission mapping
func (r *RoleRepository) UpdateRolePermission(roleID, permissionID, newPermissionID uint) error {
	var permCount int64
	_ = r.db.Table("permissions").Where("id = ? AND deleted_at IS NULL", newPermissionID).Count(&permCount)
	if permCount == 0 {
		return errors.New("new permission not found")
	}

	return r.db.Exec("UPDATE role_permissions SET permission_id = ? WHERE role_id = ? AND permission_id = ?", newPermissionID, roleID, permissionID).Error
}

// DeleteRolePermission removes a role permission mapping
func (r *RoleRepository) DeleteRolePermission(roleID, permissionID uint) error {
	res := r.db.Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("role permission mapping not found")
	}
	return nil
}

// GetUserRolesByUserID fetches user and preloads assigned roles
func (r *RoleRepository) GetUserRolesByUserID(userID uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Roles").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FetchingPermissionBasedOnRoleID retrieves all permissions for a role using preload
func (r *RoleRepository) FetchingPermissionBasedOnRoleID(roleID uint) ([]model.Permission, error) {
	var role model.Role

	err := r.db.
		Preload("Permissions").
		First(&role, roleID).Error
	if err != nil {
		return nil, err
	}

	return role.Permissions, nil
}

// FetchAllRolesPermissions retrieves all roles with associated permissions
func (r *RoleRepository) FetchAllRolesPermissions() ([]dto.RolesDTOResponse, error) {
	var roles []model.Role

	// 1. Preload permissions for all roles
	err := r.db.
		Preload("Permissions").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}

	var response []dto.RolesDTOResponse

	// 2. Assemble response DTOs
	for _, role := range roles {
		roleDTO := dto.RolesDTOResponse{
			ID:   role.ID,
			Name: role.Name,
		}

		for _, permission := range role.Permissions {
			roleDTO.Permissions = append(roleDTO.Permissions, dto.PermissionDTO{
				ID:   permission.ID,
				Name: permission.Name,
			})
		}

		response = append(response, roleDTO)
	}

	return response, nil
}


func (r *RoleRepository) GetRoleByName(
	roleName string,
) (model.Role, error) {

	var role model.Role

	roleName = strings.TrimSpace(roleName)

	err := r.db.
		Where("LOWER(name) = LOWER(?)", roleName).
		First(&role).
		Error

	if err != nil {
		return model.Role{}, err
	}

	return role, nil
}