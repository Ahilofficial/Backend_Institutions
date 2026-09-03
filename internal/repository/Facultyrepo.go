package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

// FacultyRepository handles database persistence and query operations for Faculty entities
type FacultyRepository struct {
	db *gorm.DB
}

// NewFacultyRepository initializes a new instance of FacultyRepository
func NewFacultyRepository(db *gorm.DB) *FacultyRepository {
	return &FacultyRepository{
		db: db,
	}
}

// CreateFaculty inserts a new faculty record and links the user's faculty_id
func (r *FacultyRepository) CreateFaculty(faculty *model.Faculty) error {
	// 1. Get raw sql DB connection
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	var userIDVal interface{} = nil
	if faculty.UserID > 0 {
		userIDVal = faculty.UserID
	}

	// 2. Execute insert query for faculty record
	res, err := db.Exec(
		`INSERT INTO faculties 
			(name, gender, joining_date, department_id, user_id, created_at, updated_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		faculty.Name,
		faculty.Gender,
		faculty.JoiningDate,
		faculty.DepartmentID,
		userIDVal,
		now,
		now,
		true,
	)
	if err != nil {
		return err
	}

	// 3. Extract generated faculty primary key ID
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	faculty.ID = uint(id)
	faculty.CreatedAt = now
	faculty.UpdatedAt = now
	faculty.IsActive = true

	// 4. Update user account to link newly created faculty ID
	if faculty.UserID != 0 {
		db.Exec("UPDATE users SET faculty_id = ? WHERE id = ?", faculty.ID, faculty.UserID)
	}

	return nil
}

// FetchFaculty retrieves all active faculty records
func (r *FacultyRepository) FetchFaculty() ([]model.Faculty, error) {
	// 1. Query non-deleted faculties
	var facs []model.Faculty
	err := r.db.Raw("SELECT * FROM faculties WHERE deleted_at IS NULL").Scan(&facs).Error
	if err != nil {
		return nil, err
	}

	return facs, err
}

// FetchPaidStudentsByFacultyID retrieves all non-pending students assigned to a faculty
func (r *FacultyRepository) FetchPaidStudentsByFacultyID(facultyID uint) ([]model.Student, error) {
	var students []model.Student
	err := r.db.Raw(`select * from students where faculty_id=? and deleted_at is null and pending=0`, facultyID).Scan(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

// LoginnedUserInstitutionIDRepo looks up the institution ID for an institution admin user
func (r *FacultyRepository) LoginnedUserInstitutionIDRepo(userID uint) uint {
	var luserinstid uint
	err := r.db.Raw("SELECT institution_id FROM institution_admins WHERE user_id = ? LIMIT 1", userID).Scan(&luserinstid).Error
	if err != nil {
		return 0
	}
	return luserinstid
}

// GetInstitutionIDForUserRepo looks up the institution ID that owns a faculty via department relation
func (r *FacultyRepository) GetInstitutionIDForUserRepo(facultyID uint) uint {
	var faculty model.Faculty

	// 1. Preload Department to obtain InstitutionID cleanly
	err := r.db.
		Preload("Department").
		First(&faculty, facultyID).
		Error

	if err != nil {
		return 0
	}

	return faculty.Department.InstitutionID
}

// FetchNonPaidStudentsByFacultyID retrieves pending (unpaid) students assigned to a faculty
func (r *FacultyRepository) FetchNonPaidStudentsByFacultyID(facultyID uint) ([]model.Student, error) {
	var students []model.Student
	err := r.db.Raw(`select * from students where faculty_id=? and deleted_at is null and pending=1`, facultyID).Scan(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

// FetchFacultyPaginated retrieves paginated faculty records with associated students and fee structures
func (r *FacultyRepository) FetchFacultyPaginated(page, limit int) ([]model.Faculty, int64, error) {
	var (
		facs  []model.Faculty
		total int64
	)

	// 1. Calculate total record count
	if err := r.db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	// 2. Query page with preloaded relations
	err := r.db.
		Preload("Students").
		Preload("Students.Fees").
		Preload("Students.Fees.Payments").
		Limit(limit).
		Offset(offset).
		Find(&facs).Error

	if err != nil {
		return nil, 0, err
	}

	return facs, total, nil
}

// FetchFacultyById retrieves a single faculty record by primary key ID with student relations
func (r *FacultyRepository) FetchFacultyById(id uint) (model.Faculty, error) {
	var fac model.Faculty
	err := r.db.
		Preload("Students").
		Preload("Students.Fees").
		Preload("Students.Fees.Payments").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&fac).Error
	if err != nil {
		return model.Faculty{}, err
	}

	return fac, nil
}

// FetchStudentsByFacultyID retrieves all students assigned to a faculty ID
func (r *FacultyRepository) FetchStudentsByFacultyID(facultyID uint) ([]model.Student, error) {
	var students []model.Student
	err := r.db.
		Preload("Fees").
		Preload("Fees.Payments").
		Where("faculty_id = ? AND deleted_at IS NULL", facultyID).
		Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

// DeleteFaculty soft deletes a faculty record by setting is_active = false and deleted_at timestamp
func (r *FacultyRepository) DeleteFaculty(id uint) error {
	// 1. Get database handle
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	// 2. Perform soft deletion query
	res, err := db.Exec(
		"UPDATE faculties SET is_active = ?, deleted_at = ? WHERE id = ? AND is_active = ? AND deleted_at IS NULL",
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

// UpdateFacultyById updates faculty profile fields (name, gender)
func (r *FacultyRepository) UpdateFacultyById(faculty *model.Faculty) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		"UPDATE faculties SET name = ?, gender = ?, updated_at = ? WHERE id = ?",
		faculty.Name, faculty.Gender, time.Now(), faculty.ID,
	)
	return err
}

// GetInstitutionByFacultyID looks up the institution ID associated with a faculty record
func (r *FacultyRepository) GetInstitutionByFacultyID(facultyID uint) (uint, error) {
	var faculty model.Faculty
	err := r.db.
		Preload("Department").
		Where("id = ? AND deleted_at IS NULL", facultyID).
		First(&faculty).Error
	if err != nil {
		return 0, err
	}
	return faculty.Department.InstitutionID, nil
}

// FetchByUserID retrieves faculty record matching a given user_id
func (r *FacultyRepository) FetchByUserID(userID uint) (model.Faculty, error) {
	var fac model.Faculty
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).First(&fac).Error
	return fac, err
}

// ExistsByUserID checks whether a faculty profile already exists for a user_id
func (r *FacultyRepository) ExistsByUserID(userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Faculty{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetPaidStudentsForFaculty fetches all students assigned to a faculty who have fully settled fees
func (r *FacultyRepository) GetPaidStudentsForFaculty(
	facultyID uint,
) ([]model.Student, error) {
	var students []model.Student
	err := r.db.
		Preload("Fees").
		Preload("Fees.Payments").
		Where("faculty_id = ? AND pending = ? AND deleted_at IS NULL", facultyID, false).
		Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

// GetNonPaidStudentsForFaculty fetches all students assigned to a faculty who have pending fee balances
func (r *FacultyRepository) GetNonPaidStudentsForFaculty(
	facultyID uint,
) ([]model.Student, error) {
	var students []model.Student
	err := r.db.
		Preload("Fees").
		Preload("Fees.Payments").
		Where("faculty_id = ? AND pending = ? AND deleted_at IS NULL", facultyID, true).
		Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}