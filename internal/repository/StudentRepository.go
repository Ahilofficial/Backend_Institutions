package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

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
	if student.Semester == 0 {
		student.Semester = 1
	}
	student.IsActive = true
	if err := r.db.Create(student).Error; err != nil {
		return err
	}
	if student.UserID != 0 {
		_ = r.db.Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", student.UserID).Update("student_id", student.ID).Error
	}
	return nil
}

func (r *StudentRepository) CreateStudentWithFeeTx(student *model.Student, fee *model.Fees) error {
	if student.Semester == 0 {
		student.Semester = 1
	}
	student.IsActive = true
	student.Pending = true

	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(student).Error; err != nil {
			return err
		}

		if student.UserID != 0 {
			if err := tx.Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", student.UserID).Update("student_id", student.ID).Error; err != nil {
				return err
			}
		}

		if fee != nil {
			fee.StudentID = &student.ID
			fee.Semester = student.Semester
			fee.DepartmentID = student.DepartmentID
			fee.IsActive = true
			if err := tx.Create(fee).Error; err != nil {
				return err
			}

			studentPayment := model.StudentPayment{
				StudentID:   student.ID,
				PaymentID:   fee.ID,
				Semester:    student.Semester,
				TotalAmount: fee.TotalAmount,
				Status:      "pending",
			}
			if err := tx.Create(&studentPayment).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *StudentRepository) PromoteStudentTx(studentID uint, nextSemester uint, newBaseAmount float64, newFeeAmount float64, fee *model.Fees) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&model.Student{}).Where("id = ? AND deleted_at IS NULL", studentID).Updates(map[string]interface{}{
			"semester":    nextSemester,
			"pending":     true,
			"base_amount": newBaseAmount,
			"fee_amount":  newFeeAmount,
			"updated_at":  time.Now(),
		}).Error; err != nil {
			return err
		}

		if fee != nil {
			fee.StudentID = &studentID
			fee.Semester = nextSemester
			fee.IsActive = true
			if err := tx.Create(fee).Error; err != nil {
				return err
			}

			studentPayment := model.StudentPayment{
				StudentID:   studentID,
				PaymentID:   fee.ID,
				Semester:    nextSemester,
				TotalAmount: newFeeAmount,
				Status:      "pending",
			}
			if err := tx.Create(&studentPayment).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *StudentRepository) UpdateStudentSemesterTx(
	studentID uint,
	semester uint,
	hosteller bool,
	scholarship bool,
	mq bool,
	baseAmount float64,
	feeAmount float64,
	fee *model.Fees,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		pending := true
		if fee != nil && fee.PendingAmount == 0 && fee.TotalPaid > 0 {
			pending = false
		}

		if err := tx.Model(&model.Student{}).Where("id = ? AND deleted_at IS NULL", studentID).Updates(map[string]interface{}{
			"semester":    semester,
			"hosteller":   hosteller,
			"scholarship": scholarship,
			"mq":          mq,
			"base_amount": baseAmount,
			"fee_amount":  feeAmount,
			"pending":     pending,
			"updated_at":  time.Now(),
		}).Error; err != nil {
			return err
		}

		if fee != nil {
			fee.StudentID = &studentID
			fee.Semester = semester
			fee.IsActive = true
			if fee.ID > 0 {
				if err := tx.Save(fee).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Create(fee).Error; err != nil {
					return err
				}
			}

			var sp model.StudentPayment
			err := tx.Where("student_id = ? AND semester = ? AND deleted_at IS NULL", studentID, semester).First(&sp).Error
			if err == nil && sp.ID > 0 {
				sp.TotalAmount = fee.TotalAmount
				if pending {
					sp.Status = "pending"
				}
				_ = tx.Save(&sp).Error
			} else {
				status := "pending"
				if !pending {
					status = "paid"
				}
				newSp := model.StudentPayment{
					StudentID:   studentID,
					PaymentID:   fee.ID,
					Semester:    semester,
					TotalAmount: fee.TotalAmount,
					Status:      status,
				}
				_ = tx.Create(&newSp).Error
			}
		}

		return nil
	})
}

func (r *StudentRepository) FetchByUserID(userID uint) (model.Student, error) {
	var stud model.Student
	err := r.db.Raw("SELECT * FROM students WHERE user_id = ? AND deleted_at IS NULL LIMIT 1", userID).Scan(&stud).Error
	return stud, err
}

func (r *StudentRepository) ExistsByUserID(
	userID uint,
) (bool, error) {
	var count int64
	err := r.db.Model(&model.Student{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
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
				OR students.gender LIKE ?
			)
		`,
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

func (r *StudentRepository) GetInstitutionIDForUserRepo(studentID uint) uint {
	var student model.Student
	err := r.db.
		Preload("Faculty.Department").
		Where("id = ? AND deleted_at IS NULL", studentID).
		First(&student).Error
	if err != nil {
		return 0
	}
	return student.Faculty.Department.InstitutionID
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

func (r *StudentRepository) StudentVerificationRepo(access model.StudentVerificationAccess) error {
	if access.StudentID == 0 || access.FacultyID == 0 {
		return nil
	}

	var existing model.StudentVerificationAccess
	err := r.db.Where("student_id = ? AND faculty_id = ?", access.StudentID, access.FacultyID).First(&existing).Error
	if err == nil && existing.StudentID > 0 {
		return r.db.Model(&existing).Where("student_id = ? AND faculty_id = ?", access.StudentID, access.FacultyID).Update("updated_at", time.Now()).Error
	}
	return r.db.Create(&access).Error
}

func (r *StudentRepository) HasStudentVerificationAccess(
	studentID uint,
	facultyID uint,
) (bool, error) {
	var count int64
	err := r.db.Model(&model.StudentVerificationAccess{}).
		Where("student_id = ? AND faculty_id = ?", studentID, facultyID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *StudentRepository) GetCourseDurationByFacultyIDRepo(facultyID uint) (uint, error) {
	var faculty model.Faculty
	err := r.db.
		Preload("Department").
		Where("id = ? AND deleted_at IS NULL", facultyID).
		First(&faculty).Error
	if err != nil {
		return 0, err
	}
	return faculty.Department.CourseDuration, nil
}

func (r *StudentRepository) FetchStudentByInstitution(instID uint) ([]model.Student, error) {
	var students []model.Student

	dbQuery := r.db.Model(&model.Student{}).
		Where("deleted_at IS NULL")

	if instID > 0 {
		var depts []model.Department
		_ = r.db.Where("institution_id = ? AND deleted_at IS NULL", instID).Find(&depts).Error
		var deptIDs []uint
		for _, d := range depts {
			deptIDs = append(deptIDs, d.ID)
		}
		if len(deptIDs) > 0 {
			var facs []model.Faculty
			_ = r.db.Where("department_id IN ? AND deleted_at IS NULL", deptIDs).Find(&facs).Error
			var facIDs []uint
			for _, f := range facs {
				facIDs = append(facIDs, f.ID)
			}
			if len(facIDs) > 0 {
				dbQuery = dbQuery.Where("faculty_id IN ?", facIDs)
			} else {
				return nil, nil
			}
		} else {
			return nil, nil
		}
	}

	err := dbQuery.
		Preload("Faculty").
		Preload("Fees").
		Preload("Fees.Payments").
		Find(&students).Error

	return students, err
}

func (r *StudentRepository) FetchStudentPaginatedWithInstitution(
	search string,
	page int,
	limit int,
	instID uint,
) ([]model.Student, int64, error) {
	var (
		students []model.Student
		total    int64
	)

	query := r.db.
		Model(&model.Student{}).
		Where("deleted_at IS NULL")

	if instID > 0 {
		var depts []model.Department
		_ = r.db.Where("institution_id = ? AND deleted_at IS NULL", instID).Find(&depts).Error
		var deptIDs []uint
		for _, d := range depts {
			deptIDs = append(deptIDs, d.ID)
		}
		if len(deptIDs) > 0 {
			var facs []model.Faculty
			_ = r.db.Where("department_id IN ? AND deleted_at IS NULL", deptIDs).Find(&facs).Error
			var facIDs []uint
			for _, f := range facs {
				facIDs = append(facIDs, f.ID)
			}
			if len(facIDs) > 0 {
				query = query.Where("faculty_id IN ?", facIDs)
			} else {
				return nil, 0, nil
			}
		} else {
			return nil, 0, nil
		}
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(`(name LIKE ? OR gender LIKE ?)`, searchPattern, searchPattern)
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

func (r *StudentRepository) GetStudentVerificationAccess(
	studentID uint,
	facultyID uint,
	access *model.StudentVerificationAccess,
) error {
	var student model.Student
	err := r.db.Where("id = ? AND faculty_id = ? AND deleted_at IS NULL", studentID, facultyID).First(&student).Error
	if err != nil {
		return err
	}
	access.StudentID = student.ID
	access.FacultyID = student.FacultyID
	return nil
}

func (r *StudentRepository) UpdateStudentVerified(
	userID uint,
) error {
	return r.db.
		Model(&model.Student{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Update("is_profile_verified", true).
		Error
}

func (r *StudentRepository) GetInstitutionIDByStudent(
	studentID uint,
) (uint, error) {
	var student model.Student
	err := r.db.
		Preload("Faculty.Department").
		Where("id = ? AND deleted_at IS NULL", studentID).
		First(&student).Error
	if err != nil {
		return 0, err
	}
	return student.Faculty.Department.InstitutionID, nil
}

func (r *StudentRepository) GetInstitutionByStudentID(studentID uint) (uint, error) {
	return r.GetInstitutionIDByStudent(studentID)
}

func (r *StudentRepository) CreateStudentPayment(payment *model.StudentPayment) error {
	return r.db.Create(payment).Error
}

func (r *StudentRepository) UpsertStudentPayment(payment *model.StudentPayment) error {
	var existing model.StudentPayment
	err := r.db.Where("student_id = ? AND semester = ? AND deleted_at IS NULL", payment.StudentID, payment.Semester).First(&existing).Error
	if err == nil && existing.ID > 0 {
		existing.PaymentID = payment.PaymentID
		existing.TotalAmount = payment.TotalAmount
		existing.Status = payment.Status
		existing.UpdatedAt = time.Now()
		return r.db.Save(&existing).Error
	}
	return r.db.Create(payment).Error
}

func (r *StudentRepository) UpdateStudentPendingStatus(studentID uint, pending bool) error {
	return r.db.Model(&model.Student{}).Where("id = ? AND deleted_at IS NULL", studentID).Update("pending", pending).Error
}

func (r *StudentRepository) FetchStudentPaymentsByStudentID(studentID uint) ([]model.StudentPayment, error) {
	var payments []model.StudentPayment
	err := r.db.Where("student_id = ? AND deleted_at IS NULL", studentID).Find(&payments).Error
	return payments, err
}
