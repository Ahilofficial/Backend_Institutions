package repository

import (
	"backend_institutions/internal/model"
	"strings"

	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetMenusByUser(userID uint) ([]model.Menu, error) {
	var user model.User
	if err := r.db.Preload("Roles.Menus").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return nil, err
	}

	for _, role := range user.Roles {
		if strings.EqualFold(role.Name, "super admin") || strings.EqualFold(role.Name, "super_admin") || strings.EqualFold(role.Name, "superadmin") {
			var allMenus []model.Menu
			err := r.db.Order("id ASC").Find(&allMenus).Error
			return allMenus, err
		}
	}

	menuMap := make(map[uint]model.Menu)
	for _, role := range user.Roles {
		for _, m := range role.Menus {
			menuMap[m.ID] = m
		}
	}

	var menus []model.Menu
	for _, m := range menuMap {
		menus = append(menus, m)
	}

	return menus, nil
}
