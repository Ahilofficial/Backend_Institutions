package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) CreateDepartment(department *model.Department) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	res, err := db.Exec(
		`INSERT INTO departments
			(department_name, course_duration, institution_id, created_at, updated_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?)`,
		department.DepartmentName,
		department.CourseDuration,
		department.InstitutionID,
		now,
		now,
		true,
	)
	if err != nil {
		return err
	}

	
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	department.ID = uint(id)
	department.CreatedAt = now
	department.UpdatedAt = now
	department.IsActive = true

	return nil
}

func (r *DepartmentRepository) FetchDepartment() ([]model.Department, error) {
	var depts []model.Department
	err := r.db.Raw("SELECT * FROM departments WHERE deleted_at IS NULL").Scan(&depts).Error
	if err != nil {
		return nil, err
	}

	return depts, err
}

func (r *DepartmentRepository) FetchDepartmentPaginated( page, limit int) ([]model.Department, int64, error) {
	var (
		depts []model.Department
		total int64
	)

	

	
	if err := r.db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	err := r.db.
		Preload("Faculties").
		Preload("Faculties.Students").
		Preload("Faculties.Students.Fees").
		Preload("Faculties.Students.Fees.Payments").
		Limit(limit).
		Offset(offset).
		Find(&depts).Error

	if err != nil {
		return nil, 0, err
	}

	return depts, total, nil
}

func (r *DepartmentRepository) FetchDepartmentById(id uint) (model.Department, error) {
	var dept model.Department

	err := r.db.
		Preload("Faculties").
		Preload("Faculties.Students").
		Preload("Faculties.Students.Fees").
		Preload("Faculties.Students.Fees.Payments").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&dept).Error

	if err != nil {
		return model.Department{}, err
	}

	return dept, nil
}

func (r *DepartmentRepository) DeleteDepartment(id uint) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	res, err := db.Exec(
		"UPDATE departments SET is_active = ?, deleted_at = ? WHERE id = ? AND is_active = ? AND deleted_at IS NULL",
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

func (r *DepartmentRepository) UpdateDepartmentById(department *model.Department) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		"UPDATE departments SET department_name = ?, course_duration = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		department.DepartmentName, department.CourseDuration, time.Now(), department.ID,
	)
	return err
}

func (r *DepartmentRepository) GetDepartmentFee(departmentID uint) (float64, error) {
	var feeAmount float64
	err := r.db.Raw("SELECT fee_amount FROM departments WHERE id = ? AND deleted_at IS NULL LIMIT 1", departmentID).Scan(&feeAmount).Error
	if err != nil {
		return 0, err
	}
	return feeAmount, nil
}

func (r *DepartmentRepository) UpdateDepartmentFeeAndPaymentID(departmentID uint, collegeAmount float64, hostelAmount float64, feeAmount float64, paymentID uint) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		"UPDATE departments SET college_amount = ?, hostel_amount = ?, fee_amount = ?, payment_id = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		collegeAmount, hostelAmount, feeAmount, paymentID, time.Now(), departmentID,
	)
	return err
}

func (r *DepartmentRepository) GetDepartmentByID(departmentID uint) (model.Department, error) {
	var dept model.Department
	err := r.db.Where("id = ? AND deleted_at IS NULL", departmentID).First(&dept).Error
	return dept, err
}

func (r *DepartmentRepository) GetInstitutionByDepartmentID(
	departmentID uint,
) (uint, error) {

	var institutionID uint

	err := r.db.Raw(`
		SELECT institution_id
		FROM departments
		WHERE id = ?
		LIMIT 1
	`, departmentID).Scan(&institutionID).Error

	if err != nil {
		return 0, err
	}
	return institutionID, nil
}

func (r *DepartmentRepository) GetInstitutionIDForUserRepo(id uint) uint {
	var institutionID uint
	err := r.db.Raw("SELECT institution_id FROM departments WHERE id = ? AND deleted_at IS NULL LIMIT 1", id).Scan(&institutionID).Error
	if err != nil {
		return 0
	}
	return institutionID
}