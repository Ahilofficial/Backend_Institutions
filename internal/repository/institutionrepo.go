package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

// InstitutionRepository manages database queries and modifications for Institution entities
type InstitutionRepository struct {
	db *gorm.DB
}

// NewInstitutionRepository creates a new InstitutionRepository instance
func NewInstitutionRepository(db *gorm.DB) *InstitutionRepository {
	return &InstitutionRepository{
		db: db,
	}
}

// CreateInstitution inserts a new institution record into the database
func (r *InstitutionRepository) CreateInstitution(institute *model.Institutions) error {
	// 1. Get raw database handle
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	// 2. Insert institution record
	res, err := db.Exec(
		`insert into institutions(name, institution_code,state, created_at, updated_at, is_active)values(?,?,?,?,?,?)
		`,
		institute.Name,
		institute.InstitutionCode,
		institute.State,
		now,
		now,
		true,
	)
	if err != nil {
		return err
	}

	// 3. Extract generated ID
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// 4. Update entity fields
	institute.ID = uint(id)
	institute.CreatedAt = now
	institute.UpdatedAt = now
	institute.IsActive = true

	return nil
}

// FetchInstitution retrieves all non-deleted institutions with preloaded entities
func (r *InstitutionRepository) FetchInstitution() ([]model.Institutions, error) {
	var insts []model.Institutions
	err := r.db.Preload("Departments").Preload("Faculties").Preload("Students").Preload("Fees").Where("deleted_at IS NULL").Find(&insts).Error
	if err != nil {
		return nil, err
	}
	return insts, err
}

// IsInstAdminRepo checks whether the specified user is listed as an institution admin
func (r *InstitutionRepository) IsInstAdminRepo(userID uint) bool {
	var count int64
	err := r.db.Raw("SELECT COUNT(*) FROM institution_admins WHERE user_id = ?", userID).Scan(&count).Error
	if err == nil && count > 0 {
		return true
	}

	return false
}

// GetInstitutionIDForUserRepo retrieves the institution ID assigned to an institution admin
func (r *InstitutionRepository) GetInstitutionIDForUserRepo(userID uint) uint {
	var userInstitution uint
	err := r.db.Raw("SELECT institution_id FROM institution_admins WHERE user_id = ? LIMIT 1", userID).Scan(&userInstitution).Error
	if err != nil {
		return 0
	}
	return userInstitution
}

// HasInstituteRepo returns institution ID for a user from institution_admins table
func (r *InstitutionRepository) HasInstituteRepo(userID uint, id uint) (uint, error) {
	var institutionID uint
	err := r.db.Raw("SELECT institution_id from institution_admins where user_id =?", userID).Scan(&institutionID).Error
	if err != nil {
		return 0, err
	}
	return institutionID, nil
}

// FetchInstitutionPaginated retrieves a paginated list of institutions with complete child hierarchy preloaded
func (r *InstitutionRepository) FetchInstitutionPaginated(page, limit int) ([]model.Institutions, int64, error) {
	var insts []model.Institutions
	var total int64
	offset := (page - 1) * limit

	// 1. Preload child relationships with pagination
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

// FetchInstitutionById retrieves an institution by ID with full relations preloaded
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

// DeleteInstitution soft deletes an institution
func (r *InstitutionRepository) DeleteInstitution(id uint) error {
	// 1. Get database handle
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	// 2. Perform soft delete update
	res, err := db.Exec(
		"UPDATE institutions SET is_active = ?, deleted_at = ? WHERE id = ? AND is_active = ? AND deleted_at IS NULL",
		false, time.Now(), id, true,
	)
	if err != nil {
		return err
	}

	// 3. Verify affected rows
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}
	return nil
}

// UpdateInstitution updates institution name, code, and state
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

// GetInstitutionIDByUserID retrieves institution ID for user
func (r *InstitutionRepository) GetInstitutionIDByUserID(userID uint) (uint, error) {
	var institutionID uint
	err := r.db.Raw("SELECT institution_id FROM institution_admins WHERE user_id = ?", userID).Scan(&institutionID).Error
	if err != nil {
		return 0, err
	}
	return institutionID, nil
}
