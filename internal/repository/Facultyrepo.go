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

func(r *FacultyRepository)FetchPaidStudentsByFacultyID(facultyID uint) ([]model.Student, error) {
	var students []model.Student
	err := r.db.Raw(`select * from students where faculty_id=? and deleted_at is null and pending=0`,facultyID).Scan(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (r *FacultyRepository) LoginnedUserInstitutionIDRepo(userID uint) uint {
	var luserinstid uint
	err := r.db.Raw("SELECT institution_id FROM institution_admins WHERE user_id = ? LIMIT 1", userID).Scan(&luserinstid).Error
	if err != nil {
		return 0
	}
	return luserinstid
}
func (r *FacultyRepository) GetInstitutionIDForUserRepo(facultyID uint) uint {
	var faculty model.Faculty

	err := r.db.
		Preload("Department").
		First(&faculty, facultyID).
		Error

	if err != nil {
		return 0
	}

	return faculty.Department.InstitutionID
}

func(r *FacultyRepository) FetchNonPaidStudentsByFacultyID(facultyID uint) ([]model.Student, error) {
	var students []model.Student
	err := r.db.Raw(`select * from students where faculty_id=? and deleted_at is null and pending=1`,facultyID).Scan(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (r *FacultyRepository) FetchFacultyPaginated(page, limit int) ([]model.Faculty, int64, error) {
	var (
		facs  []model.Faculty
		total int64
	)


	

	if err := r.db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

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

func (r *FacultyRepository) GetDepartmentFeeByFacultyID(facultyID uint) (float64, error) {
	var feeAmount float64

	err := r.db.Raw(`
		SELECT d.fee_amount
		FROM faculties f
		JOIN departments d ON f.department_id = d.id
		WHERE f.id = ?
		  AND f.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		LIMIT 1
	`, facultyID).Scan(&feeAmount).Error

	if err != nil {
		return 0, err
	}

	return feeAmount, nil
}

func (r *FacultyRepository) GetDepartmentFeeAndPaymentIDByFacultyID(facultyID uint) (float64, float64, float64, uint, error) {
	type Result struct {
		CollegeAmount float64 `gorm:"column:college_amount"`
		HostelAmount  float64 `gorm:"column:hostel_amount"`
		FeeAmount     float64 `gorm:"column:fee_amount"`
		PaymentID     uint    `gorm:"column:payment_id"`
	}
	var res Result

	err := r.db.Raw(`
		SELECT d.college_amount, d.hostel_amount, d.fee_amount, d.payment_id
		FROM faculties f
		JOIN departments d ON f.department_id = d.id
		WHERE f.id = ?
		  AND f.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		LIMIT 1
	`, facultyID).Scan(&res).Error

	if err != nil {
		return 0, 0, 0, 0, err
	}

	return res.CollegeAmount, res.HostelAmount, res.FeeAmount, res.PaymentID, nil
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


func (r *UserRepository) GetFacultyIDForUser(userID uint) (uint, error) {
	var facultyID uint

	err := r.db.Raw(
		`SELECT faculty_id
		 FROM users
		 WHERE id = ?`,
		userID,
	).Scan(&facultyID).Error

	if err != nil {
		return 0, err
	}

	if facultyID == 0 {
		return 0, errors.New("faculty profile not found for this user")
	}

	return facultyID, nil
}

func (r *FacultyRepository) GetPaidStudentsForFaculty(
	facultyID uint,
) ([]model.Student, error) {

	var students []model.Student

	err := r.db.Raw(`
		SELECT DISTINCT s.*
		FROM students s
		JOIN payments p ON p.student_id = s.id
		WHERE s.faculty_id = ?
		  AND p.payment_status = 'paid'
	`, facultyID).Scan(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}

func (r *FacultyRepository) GetNonPaidStudentsForFaculty(
	facultyID uint,
) ([]model.Student, error) {

	var students []model.Student

	err := r.db.Raw(`
		SELECT s.*
		FROM students s
		WHERE s.faculty_id = ?
		AND NOT EXISTS (
			SELECT 1
			FROM payments p
			WHERE p.student_id = s.id
			AND p.payment_status = 'paid'
		)
	`, facultyID).Scan(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}