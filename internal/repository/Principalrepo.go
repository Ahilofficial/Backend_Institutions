package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type PrincipalRepository struct {
	db *gorm.DB
}

func NewPrincipalRepository(db *gorm.DB) *PrincipalRepository {
	return &PrincipalRepository{
		db: db,
	}
}

func (r *PrincipalRepository) CreatePrincipal(principal *model.Principal) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	res, err := db.Exec(
		`INSERT INTO principals
			(name, gender, joining_date, institution_id, user_id, created_at, updated_at, is_active)
		SELECT ?, ?, ?, id, ?, ?, ?, ?
		FROM institutions
		WHERE id = ?
		  AND deleted_at IS NULL
		  AND is_active = true
		  AND NOT EXISTS (
			  SELECT 1
			  FROM principals
			  WHERE user_id = ?
			    AND user_id > 0
			    AND deleted_at IS NULL
		  )`,
		principal.Name,
		principal.Gender,
		principal.JoiningDate,
		principal.UserID,
		now,
		now,
		true,
		principal.InstitutionID,
		principal.UserID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("principal profile already exists for this user, or parent institution is inactive/invalid")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	principal.ID = uint(id)
	principal.CreatedAt = now
	principal.UpdatedAt = now
	principal.IsActive = true

	if principal.UserID != 0 {
		db.Exec("UPDATE users SET principal_id = ? WHERE id = ?", principal.ID, principal.UserID)
	}

	return nil
}

func (r *PrincipalRepository) FetchPrincipal() ([]model.Principal, error) {
	var prs []model.Principal
	err := r.db.Raw("SELECT * FROM principals WHERE deleted_at IS NULL").Scan(&prs).Error
	if err != nil {
		return nil, err
	}

	return prs, err
}

func (r *PrincipalRepository) FetchPrincipalPaginated(search string, page, limit int) ([]model.Principal, int64, error) {
	var (
		prs   []model.Principal
		total int64
	)

	query := r.db.Model(&model.Principal{})

	if search != "" {
		search = "%" + search + "%"
		query = query.Where(`
			name LIKE ? OR
			gender LIKE ? OR
			joining_date LIKE ?
		`, search, search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	err := query.
		Limit(limit).
		Offset(offset).
		Find(&prs).Error

	if err != nil {
		return nil, 0, err
	}

	return prs, total, nil
}

func (r *PrincipalRepository) FetchPrincipalById(id uint) (model.Principal, error) {
	var pr model.Principal
	err := r.db.
		Where("id = ? AND deleted_at IS NULL", id).
		First(&pr).Error
	if err != nil {
		return model.Principal{}, err
	}

	return pr, nil
}

func (r *PrincipalRepository) FetchPrincipalDeleted() ([]model.Principal, error) {
	var prs []model.Principal
	err := r.db.Raw("SELECT * FROM principals WHERE deleted_at IS NOT NULL").Scan(&prs).Error
	if err != nil {
		return nil, err
	}
	return prs, err
}

func (r *PrincipalRepository) GetActivePrincipal() (model.Principal, error) {
	var prs []model.Principal
	err := r.db.Raw("SELECT * FROM principals WHERE is_active = ? AND deleted_at IS NULL LIMIT 1", true).Scan(&prs).Error
	if err != nil {
		return model.Principal{}, err
	}
	if len(prs) == 0 {
		return model.Principal{}, gorm.ErrRecordNotFound
	}

	return prs[0], nil
}

func (r *PrincipalRepository) GetInactivePrincipal() (model.Principal, error) {
	var prs []model.Principal
	err := r.db.Raw("SELECT * FROM principals WHERE is_active = ? AND deleted_at IS NULL LIMIT 1", false).Scan(&prs).Error
	if err != nil {
		return model.Principal{}, err
	}
	if len(prs) == 0 {
		return model.Principal{}, gorm.ErrRecordNotFound
	}

	return prs[0], nil
}

func (r *PrincipalRepository) DeletePrincipal(id uint) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	res, err := db.Exec(
		"UPDATE principals SET is_active = ?, deleted_at = ? WHERE id = ? AND is_active = ? AND deleted_at IS NULL",
		false, time.Now(), id, true,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}
	return nil
}

func (r *PrincipalRepository) UpdatePrincipalById(principal *model.Principal) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		"UPDATE principals SET name = ?, gender = ?, updated_at = ? WHERE id = ?",
		principal.Name, principal.Gender, time.Now(), principal.ID,
	)
	return err
}

func (r *PrincipalRepository) GetInstitutionByPrincipalID(principalID uint) (uint, error) {
	var institutionID uint

	err := r.db.Raw(`
		SELECT d.institution_id
		FROM principals pr
		JOIN departments d ON pr.department_id = d.id
		WHERE pr.id = ?
		  AND pr.deleted_at IS NULL
		  AND d.deleted_at IS NULL
	`, principalID).Scan(&institutionID).Error

	if err != nil {
		return 0, err
	}

	return institutionID, nil
}

func (r *PrincipalRepository) GetInstitutionIDByPrincipal(principalID uint) (uint, error) {
	var institutionID uint

	err := r.db.Raw(`
		SELECT institution_id
		FROM principals
		WHERE id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, principalID).Scan(&institutionID).Error

	if err != nil {
		return 0, err
	}

	return institutionID, nil
}

func (r *PrincipalRepository) FetchByUserID(userID uint) (model.Principal, error) {
	var pr model.Principal
	err := r.db.Raw("SELECT * FROM principals WHERE user_id = ? AND deleted_at IS NULL LIMIT 1", userID).Scan(&pr).Error
	return pr, err
}

func (r *PrincipalRepository) ExistsByUserID(userID uint) (bool, error) {
	var exists bool

	result := r.db.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM principals
			WHERE user_id = ?
		)
	`, userID).Scan(&exists)

	if result.Error != nil {
		return false, result.Error
	}

	return exists, nil
}
