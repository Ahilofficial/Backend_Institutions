package repository

import (
	"errors"
	"strings"
	"time"

	"backend_institutions/internal/model"

	"gorm.io/gorm"
)

type StudentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{
		db: db,
	}
}



func (r *StudentRepository) CreateStudent(
	student *model.Student,
) error {

	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	res, err := db.Exec(
		`INSERT INTO students
			(name, gender, faculty_id, user_id, created_at, updated_at, is_active)
		SELECT ?, ?, id, ?, ?, ?, ?
		FROM faculties
		WHERE id = ?
		  AND deleted_at IS NULL
		  AND is_active = true
		  AND NOT EXISTS (
			  SELECT 1
			  FROM students
			  WHERE user_id = ?
			    AND user_id > 0
			    AND deleted_at IS NULL
		  )`,
		student.Name,
		student.Gender,
		student.UserID,
		now,
		now,
		true,
		student.FacultyID,
		student.UserID,
	)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New(
			"student profile already exists for this user, or assigned faculty is inactive/invalid",
		)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	student.ID = uint(id)
	student.CreatedAt = now
	student.UpdatedAt = now
	student.IsActive = true

	if student.UserID != 0 {
		db.Exec("UPDATE users SET student_id = ? WHERE id = ?", student.ID, student.UserID)
	}

	return nil
}



func (r *StudentRepository) FetchByUserID(userID uint) (model.Student, error) {
	var stud model.Student
	err := r.db.Raw("SELECT * FROM students WHERE user_id = ? AND deleted_at IS NULL LIMIT 1", userID).Scan(&stud).Error
	return stud, err
}

func (r *StudentRepository) ExistsByUserID(
	userID uint,
) (bool, error) {

	var exists bool

	result := r.db.Raw(`
		SELECT EXISTS(
			SELECT 1
			FROM students
			WHERE user_id = ?
			  AND deleted_at IS NULL
		)
	`, userID).Scan(&exists)

	if result.Error != nil {
		return false, result.Error
	}

	return exists, nil
}



func (r *StudentRepository) FetchStudent() ([]model.Student, error) {

	var students []model.Student

	err := r.db.
		Where("deleted_at IS NULL").
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}



func (r *StudentRepository) FetchStudentPaginated(
	search string,
	page int,
	limit int,
) ([]model.Student, int64, error) {

	var (
		students []model.Student
		total    int64
	)

	query := r.db.
		Model(&model.Student{}).
		Where("students.deleted_at IS NULL")

	if search != "" {

		searchPattern := "%" + search + "%"

		query = query.Where(`
			(
				students.name LIKE ?
				OR students.email LIKE ?
				OR students.gender LIKE ?
			)
		`,
			searchPattern,
			searchPattern,
			searchPattern,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	err := query.
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		Limit(limit).
		Offset(offset).
		Find(&students).Error

	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}



func (r *StudentRepository) FetchStudentById(
	id uint,
) (model.Student, error) {

	var student model.Student

	err := r.db.
		Where("id = ? AND deleted_at IS NULL", id).
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		First(&student).Error

	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}



func (r *StudentRepository) FetchStudentDeleted() ([]model.Student, error) {

	var students []model.Student

	err := r.db.
		Unscoped().
		Where("deleted_at IS NOT NULL").
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}



func (r *StudentRepository) GetActiveStudent() (model.Student, error) {

	var student model.Student

	err := r.db.
		Where("is_active = ? AND deleted_at IS NULL", true).
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		First(&student).Error

	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}



func (r *StudentRepository) GetInactiveStudent() (model.Student, error) {

	var student model.Student

	err := r.db.
		Where("is_active = ? AND deleted_at IS NULL", false).
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		First(&student).Error

	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}



func (r *StudentRepository) DeleteStudent(id uint) error {

	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	res, err := db.Exec(
		`UPDATE students
		 SET is_active = ?,
		     deleted_at = ?
		 WHERE id = ?
		   AND is_active = ?
		   AND deleted_at IS NULL`,
		false,
		now,
		id,
		true,
	)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New(
			"record not found or already deleted",
		)
	}

	return nil
}



func (r *StudentRepository) UpdateStudentById(
	student *model.Student,
) error {

	db, err := r.db.DB()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`UPDATE students
		 SET name = ?,
		     gender = ?,
		     updated_at = ?
		 WHERE id = ?
		   AND deleted_at IS NULL`,
		student.Name,
		student.Gender,
		time.Now(),
		student.ID,
	)

	return err
}



func (r *StudentRepository) FetchStudentsByPaymentMonth(
	month string,
) ([]model.Student, error) {

	var students []model.Student

	err := r.db.
		Model(&model.Student{}).
		Joins("JOIN fees ON fees.student_id = students.id").
		Joins("JOIN payments ON payments.fee_id = fees.id").
		Where(`
			students.deleted_at IS NULL
			AND LOWER(payments.month) = LOWER(?)
		`, month).
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		Distinct().
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}


func (r *StudentRepository) FetchPaidStudents() ([]model.Student, error) {
	return r.FetchPaidStudentsByMonth(0, 0, "")
}

func (r *StudentRepository) FetchPaidStudentsByMonth(instID uint, facultyID uint, month string) ([]model.Student, error) {
	var students []model.Student

	dbQuery := r.db.Model(&model.Student{}).
		Joins("JOIN faculties f ON f.id = students.faculty_id").
		Joins("JOIN departments d ON d.id = f.department_id").
		Joins("JOIN fees ON fees.student_id = students.id").
		Joins("JOIN payments ON payments.fee_id = fees.id").
		Where("students.deleted_at IS NULL AND payments.deleted_at IS NULL AND payments.amount_paid > 0")

	if strings.TrimSpace(month) != "" {
		dbQuery = dbQuery.Where("LOWER(payments.month) = LOWER(?)", strings.TrimSpace(month))
	}

	if facultyID > 0 {
		dbQuery = dbQuery.Where("students.faculty_id = ?", facultyID)
	} else if instID > 0 {
		dbQuery = dbQuery.Where("d.institution_id = ?", instID)
	}

	err := dbQuery.
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		Distinct().
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}

func (r *StudentRepository) FetchNotPaidStudents() ([]model.Student, error) {
	return r.FetchNotPaidStudentsByMonth(0, 0, "")
}

func (r *StudentRepository) FetchNotPaidStudentsByMonth(instID uint, facultyID uint, month string) ([]model.Student, error) {
	var students []model.Student

	m := strings.TrimSpace(month)

	dbQuery := r.db.Model(&model.Student{}).
		Joins("JOIN faculties f ON f.id = students.faculty_id").
		Joins("JOIN departments d ON d.id = f.department_id").
		Where("students.deleted_at IS NULL")

	if m != "" {
		dbQuery = dbQuery.Where("students.id NOT IN (SELECT f.student_id FROM fees f JOIN payments p ON p.fee_id = f.id WHERE LOWER(p.month) = LOWER(?) AND p.deleted_at IS NULL AND p.amount_paid > 0)", m)
	} else {
		dbQuery = dbQuery.Where("students.id NOT IN (SELECT student_id FROM fees WHERE total_amount = total_paid)")
	}

	if facultyID > 0 {
		dbQuery = dbQuery.Where("students.faculty_id = ?", facultyID)
	} else if instID > 0 {
		dbQuery = dbQuery.Where("d.institution_id = ?", instID)
	}

	err := dbQuery.
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		Distinct().
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	return students, nil
}


func (r *StudentRepository) GetInstitutionIDByStudent(
	studentID uint,
) (uint, error) {

	var institutionID uint

	err := r.db.Raw(`
		SELECT d.institution_id
		FROM students s
		JOIN faculties f
			ON s.faculty_id = f.id
		JOIN departments d
			ON f.department_id = d.id
		WHERE s.id = ?
		  AND s.deleted_at IS NULL
	`, studentID).Scan(&institutionID).Error

	if err != nil {
		return 0, err
	}

	return institutionID, nil
}

func (r *StudentRepository) GetInstitutionByStudentID(studentID uint) (uint, error) {
	return r.GetInstitutionIDByStudent(studentID)
}
