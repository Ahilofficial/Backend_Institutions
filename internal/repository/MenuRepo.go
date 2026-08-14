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
	var menus []model.Menu

	query := `
	SELECT
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
