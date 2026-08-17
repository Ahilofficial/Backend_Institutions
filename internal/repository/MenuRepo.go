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
		superAdminMenus := []model.Menu{
			{ID: 1, Name: "Dashboard", Route: "/dashboard", Icon: "dashboard"},
			{ID: 2, Name: "My Profile", Route: "/profile", Icon: "person"},
			{ID: 3, Name: "My Student Details", Route: "/student-details", Icon: "groups"},
			{ID: 4, Name: "My Institution", Route: "/institution", Icon: "account_balance"},
			{ID: 5, Name: "My Department", Route: "/department", Icon: "domain"},
			{ID: 6, Name: "My Faculty", Route: "/faculty", Icon: "school"},
		}
		return superAdminMenus, nil
	}

	var isFaculty bool
	_ = database.DB.Raw(`
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur 
			JOIN roles r ON r.id = ur.role_id 
			WHERE ur.user_id = ? AND LOWER(r.name) = 'faculty'
		)
	`, userID).Scan(&isFaculty)

	if isFaculty {
		facultyMenus := []model.Menu{
			{ID: 1, Name: "Dashboard", Route: "/dashboard", Icon: "dashboard"},
			{ID: 2, Name: "My Profile", Route: "/profile", Icon: "person"},
			{ID: 3, Name: "My Students", Route: "/students", Icon: "groups"},
		}
		return facultyMenus, nil
	}

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
