package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type InstitutionRepository struct {
	db *gorm.DB
}

func NewInstitutionRepository(db *gorm.DB) *InstitutionRepository {
	return &InstitutionRepository{
		db: db,
	}
}

func (r *InstitutionRepository) CreateInstitution(institute *model.Institutions) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()
	// model.Institutions

	res, err := db.Exec(
		`insert into institutions(name, institution_code,state, created_at, updated_at, is_active)values(?,?,?,?,?,?)
		`,
		institute.Name,
		institute.InstitutionCode,
		institute.State,
		now,
		now,
		true,
		// institute.Name,
		// institute.InstitutionCode,
	)
	if err != nil {
		return err
	}


	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	institute.ID = uint(id)
	institute.CreatedAt = now
	institute.UpdatedAt = now
	institute.IsActive = true

	return nil
}

func (r *InstitutionRepository) FetchInstitution() ([]model.Institutions, error) {
	var insts []model.Institutions
	err := r.db.Preload("Departments").Preload("Faculties").Preload("Students").Preload("Fees").Where("deleted_at IS NULL").Find(&insts).Error
	if err != nil {
		return nil, err
	}
	return insts, err
}

func (r *InstitutionRepository) IsInstAdminRepo(userID uint) bool {
	
	var count int64
	err := r.db.Raw("SELECT COUNT(*) FROM institution_admins WHERE user_id = ?", userID).Scan(&count).Error
	if err == nil && count > 0 {
		return true
	}

	return false
}

func (r *InstitutionRepository) GetInstitutionIDForUserRepo(userID uint) uint {
	
	var userInstitution uint
	err := r.db.Raw("SELECT institution_id FROM institution_admins WHERE user_id = ? LIMIT 1", userID).Scan(&userInstitution).Error
	if err != nil {
		return 0
	}
	return userInstitution
}
func (r *InstitutionRepository) HasInstituteRepo(userID uint, id uint) (uint, error) {
	var institutionID uint
	err := r.db.Raw("SELECT institution_id from institution_admins where user_id =?", userID).Scan(&institutionID).Error
	if err != nil {
		return 0, err
	}
	return institutionID, nil
}
func (r *InstitutionRepository) FetchInstitutionPaginated( page, limit int) ([]model.Institutions, int64, error) {
	
	var insts []model.Institutions
	var total int64
	offset := (page - 1) * limit

	err := r.db.
		Preload("Departments").
		Preload("Departments.Faculties").
		Preload("Departments.Faculties.Students").
		Preload("Departments.Faculties.Students.Fees").
		Preload("Departments.Faculties.Students.Fees.Payments").
		Limit(limit).
		Offset(offset).
		Find(&insts).Error

	if err != nil {
		return nil, 0, err
	}

	return insts, total, nil
}

func (r *InstitutionRepository) FetchInstitutionById(id uint) (model.Institutions, error) {
	var inst model.Institutions

	err := r.db.
		Preload("Departments").
		Preload("Departments.Faculties").
		Preload("Departments.Faculties.Students").
		Preload("Departments.Faculties.Students.Fees").
		Preload("Departments.Faculties.Students.Fees.Payments").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&inst).Error

	if err != nil {
		return model.Institutions{}, err
	}

	return inst, nil
}


func (r *InstitutionRepository) DeleteInstitution(id uint) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	res, err := db.Exec(
		"UPDATE institutions SET is_active = ?, deleted_at = ? WHERE id = ? AND is_active = ? AND deleted_at IS NULL",
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

func (r *InstitutionRepository) UpdateInstitution(institute *model.Institutions) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		"UPDATE institutions SET name = ?, institution_code = ?, state = ?, updated_at = ? WHERE id = ?",
		institute.Name, institute.InstitutionCode, institute.State, time.Now(), institute.ID,
	)
	return err
}

func (r *InstitutionRepository) GetInstitutionIDByUserID(userID uint) (uint, error) {
	var institutionID uint
	err := r.db.Raw("SELECT institution_id FROM institution_admins WHERE user_id = ?", userID).Scan(&institutionID).Error
	if err != nil {
		return 0, err
	}
	return institutionID, nil
}
