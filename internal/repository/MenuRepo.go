package repository

import (
	"backend_institutions/internal/model"
	"strings"

	"gorm.io/gorm"
)

// MenuRepository handles querying dynamic sidebar/navigation menus for users based on their assigned roles
type MenuRepository struct {
	db *gorm.DB
}

// NewMenuRepository instantiates a new MenuRepository
func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// GetMenusByUser fetches all accessible navigation menus for a user
func (r *MenuRepository) GetMenusByUser(userID uint) ([]model.Menu, error) {
	// 1. Preload user roles and their associated menus
	var user model.User
	if err := r.db.Preload("Roles.Menus").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return nil, err
	}

	// 2. Grant super admins access to all system menus
	for _, role := range user.Roles {
		if strings.EqualFold(role.Name, "super admin") || strings.EqualFold(role.Name, "super_admin") || strings.EqualFold(role.Name, "superadmin") {
			var allMenus []model.Menu
			err := r.db.Order("id ASC").Find(&allMenus).Error
			return allMenus, err
		}
	}

	// 3. Deduplicate menus across all assigned user roles
	menuMap := make(map[uint]model.Menu)
	for _, role := range user.Roles {
		for _, m := range role.Menus {
			menuMap[m.ID] = m
		}
	}

	// 4. Flatten deduplicated map into a slice
	var menus []model.Menu
	for _, m := range menuMap {
		menus = append(menus, m)
	}

	return menus, nil
}
