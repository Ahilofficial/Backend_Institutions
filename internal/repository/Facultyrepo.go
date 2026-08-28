package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type FacultyRepository struct {
	db *gorm.DB
}

func NewFacultyRepository(db *gorm.DB) *FacultyRepository {
	return &FacultyRepository{
		db: db,
	}
}

func (r *FacultyRepository) CreateFaculty(faculty *model.Faculty) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	var userIDVal interface{} = nil
	if faculty.UserID > 0 {
		userIDVal = faculty.UserID
	}

	res, err := db.Exec(
		`INSERT INTO faculties
			(name, gender, joining_date, department_id, user_id, created_at, updated_at, is_active)
		SELECT ?, ?, ?, id, ?, ?, ?, ?
		FROM departments
		WHERE id = ?
		  AND deleted_at IS NULL
		  AND is_active = true
		  AND (? IS NULL OR NOT EXISTS (
			  SELECT 1
			  FROM faculties
			  WHERE user_id = ?
			    AND deleted_at IS NULL
		  ))`,
		faculty.Name,
		faculty.Gender,
		faculty.JoiningDate,
		userIDVal,
		now,
		now,
		true,
		faculty.DepartmentID,
		userIDVal,
		userIDVal,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("faculty profile already exists for this user, or parent department is inactive/invalid")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	faculty.ID = uint(id)
	faculty.CreatedAt = now
	faculty.UpdatedAt = now
	faculty.IsActive = true

	if faculty.UserID != 0 {
		db.Exec("UPDATE users SET faculty_id = ? WHERE id = ?", faculty.ID, faculty.UserID)
	}

	return nil
}

func (r *FacultyRepository) FetchFaculty() ([]model.Faculty, error) {
	var facs []model.Faculty
	err := r.db.Raw("SELECT * FROM faculties WHERE deleted_at IS NULL").Scan(&facs).Error
	if err != nil {
		return nil, err
	}

	return facs, err
}

func (r *FacultyRepository) LoginnedUserInstitutionIDRepo(userID uint) uint {
	var luserinstid uint
	err := r.db.Raw("SELECT institution_id FROM institution_admins WHERE user_id = ? LIMIT 1", userID).Scan(&luserinstid).Error
	if err != nil {
		return 0
	}
	return luserinstid
}
func (r *FacultyRepository)GetInstitutionIDForUserRepo(FacultyID uint)(uint){
	var faculty_inst_id uint
	err:=r.db.Raw(`SELECT d.institution_id
		FROM faculties f
		JOIN departments d ON d.id = f.department_id
		WHERE f.id = ?
`).Scan(&faculty_inst_id).Error
	if err != nil {
		return 0
	}
	return faculty_inst_id
}

func (r *FacultyRepository) FetchFacultyPaginated(search string, page, limit int) ([]model.Faculty, int64, error) {
	var (
		facs  []model.Faculty
		total int64
	)

	query := r.db.Model(&model.Faculty{})

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

func (r *FacultyRepository) FetchFacultyDeleted() ([]model.Faculty, error) {
	var facs []model.Faculty
	err := r.db.Raw("SELECT * FROM faculties WHERE deleted_at IS NOT NULL").Scan(&facs).Error
	if err != nil {
		return nil, err
	}
	return facs, err
}

func (r *FacultyRepository) GetActiveFaculty() (model.Faculty, error) {
	var facs []model.Faculty
	err := r.db.Raw("SELECT * FROM faculties WHERE is_active = ? AND deleted_at IS NULL LIMIT 1", true).Scan(&facs).Error
	if err != nil {
		return model.Faculty{}, err
	}
	if len(facs) == 0 {
		return model.Faculty{}, gorm.ErrRecordNotFound
	}

	return facs[0], nil
}

func (r *FacultyRepository) GetInactiveFaculty() (model.Faculty, error) {
	var facs []model.Faculty
	err := r.db.Raw("SELECT * FROM faculties WHERE is_active = ? AND deleted_at IS NULL LIMIT 1", false).Scan(&facs).Error
	if err != nil {
		return model.Faculty{}, err
	}
	if len(facs) == 0 {
		return model.Faculty{}, gorm.ErrRecordNotFound
	}

	return facs[0], nil
}

func (r *FacultyRepository) DeleteFaculty(id uint) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	res, err := db.Exec(
		"UPDATE faculties SET is_active = ?, deleted_at = ? WHERE id = ? AND is_active = ? AND deleted_at IS NULL",
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
func (r *FacultyRepository) GetInstitutionByFacultyID(facultyID uint) (uint, error) {
	var institutionID uint

	err := r.db.Raw(`
		SELECT d.institution_id
		FROM faculties f
		JOIN departments d ON f.department_id = d.id
		WHERE f.id = ?
		  AND f.deleted_at IS NULL
		  AND d.deleted_at IS NULL
	`, facultyID).Scan(&institutionID).Error

	if err != nil {
		return 0, err
	}

	return institutionID, nil
}

func (r *FacultyRepository) FetchByUserID(userID uint) (model.Faculty, error) {
	var fac model.Faculty
	err := r.db.Raw("SELECT * FROM faculties WHERE user_id = ? AND deleted_at IS NULL LIMIT 1", userID).Scan(&fac).Error
	return fac, err
}

func (r *FacultyRepository) ExistsByUserID(userID uint) (bool, error) {

	var exists bool

	result := r.db.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM faculties
			WHERE user_id = ?
		)
	`, userID).Scan(&exists)

	if result.Error != nil {
		return false, result.Error
	}

	return exists, nil
}
