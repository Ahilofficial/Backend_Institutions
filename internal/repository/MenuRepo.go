package repository

import (
	"backend_institutions/internal/database"
	"backend_institutions/internal/model"

	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetMenusByUser(userID uint) ([]model.Menu, error) {
	var isSuperAdmin bool
	_ = database.DB.Raw(`
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur 
			JOIN roles r ON r.id = ur.role_id 
			WHERE ur.user_id = ? AND LOWER(r.name) IN ('super admin', 'super_admin', 'superadmin')
		)
	`, userID).Scan(&isSuperAdmin)

	if isSuperAdmin {
		var menus []model.Menu
		err := database.DB.Raw("SELECT id, name, route, icon, parent_id FROM menus ORDER BY id").Scan(&menus).Error
		return menus, err
	}

	var menus []model.Menu

	query := `
	SELECT DISTINCT
		m.id,
		m.name,
		m.route,
		m.icon,
		m.parent_id
	FROM menus m
	INNER JOIN role_menus rm
		ON rm.menu_id = m.id
	INNER JOIN user_roles ur
		ON ur.role_id = rm.role_id
	WHERE ur.user_id = ?
	ORDER BY
		CASE
			WHEN m.parent_id IS NULL THEN m.id
			ELSE m.parent_id
		END,
		m.id
	`

	err := database.DB.Raw(query, userID).Scan(&menus).Error

	if err != nil {
		return nil, err
	}

	return menus, nil
}
