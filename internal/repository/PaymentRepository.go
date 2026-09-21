package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) CreateDepartmentPayment(p *model.DepartmentPayment) error {
	return r.db.Create(p).Error
}

func (r *PaymentRepository) GetDepartmentPaymentBySemester(deptID uint, semester uint) (*model.DepartmentPayment, error) {
	var p model.DepartmentPayment

	result := r.db.Raw(`
		SELECT *
		FROM department_payments
		WHERE department_id = ?
		  AND semester = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, deptID, semester).Scan(&p)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &p, nil
}

func (r *PaymentRepository) GetDepartmentPaymentByID(id uint) (*model.DepartmentPayment, error) {
	var p model.DepartmentPayment

	result := r.db.Raw(`
		SELECT *
		FROM department_payments
		WHERE id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, id).Scan(&p)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var department model.Department

	result = r.db.Raw(`
		SELECT *
		FROM departments
		WHERE id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, p.DepartmentID).Scan(&department)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	p.Department = &department

	return &p, nil
}

func (r *PaymentRepository) GetDepartmentPaymentsByDeptID(deptID uint) ([]model.DepartmentPayment, error) {
	var list []model.DepartmentPayment

	result := r.db.Raw(`
		SELECT *
		FROM department_payments
		WHERE department_id = ?
		  AND deleted_at IS NULL
		ORDER BY semester ASC
	`, deptID).Scan(&list)

	if result.Error != nil {
		return nil, result.Error
	}

	return list, nil
}

func (r *PaymentRepository) UpdateDepartmentPayment(p *model.DepartmentPayment) error {
	result := r.db.Exec(`
		UPDATE department_payments
		SET college_amount = ?,
			hostel_amount = ?,
			total_amount = ?,
			updated_at = ?
		WHERE id = ?
		  AND deleted_at IS NULL
	`,
		p.CollegeAmount,
		p.HostelAmount,
		p.TotalAmount,
		time.Now(),
		p.ID,
	)

	return result.Error
}
func (r *PaymentRepository) DeleteDepartmentPayment(id uint) error {
	res := r.db.Exec(`
		UPDATE department_payments
		SET deleted_at = ?
		WHERE id = ?
		  AND deleted_at IS NULL
	`, time.Now(), id)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("payment record not found or already deleted")
	}

	return nil
}

func (r *PaymentRepository) CreateStudentPayment(p *model.StudentPayment) error {
	now := time.Now()

	result := r.db.Exec(`
		INSERT INTO student_payments (
			student_id,
			department_payment_id,
			amount_paid,
			payment_mode,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		p.StudentID,
		p.DepartmentPaymentID,
		p.AmountPaid,
		p.PaymentMode,
		now,
		now,
	)

	return result.Error
}

func (r *PaymentRepository) GetStudentPaymentsByStudentID(studentID uint) ([]model.StudentPayment, error) {
	var payments []model.StudentPayment

	result := r.db.Raw(`
		SELECT *
		FROM student_payments
		WHERE student_id = ?
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, studentID).Scan(&payments)

	if result.Error != nil {
		return nil, result.Error
	}

	return payments, nil
}

func (r *PaymentRepository) GetStudentByID(studentID uint) (*model.Student, error) {
    var s model.Student

    result := r.db.Raw(`
        SELECT *
        FROM students
        WHERE id = ?
          AND deleted_at IS NULL
        LIMIT 1
    `, studentID).Scan(&s)

    if result.Error != nil {
        return nil, result.Error
    }

    if result.RowsAffected == 0 {
        return nil, gorm.ErrRecordNotFound
    }

    return &s, nil
}

func (r *PaymentRepository) UpdateStudentPaymentStatus(student *model.Student) error {
	result := r.db.Exec(`
		UPDATE students
		SET paid_amount = ?,
			pending = ?,
			updated_at = ?
		WHERE id = ?
		  AND deleted_at IS NULL
	`,
		student.PaidAmount,
		student.Pending,
		time.Now(),
		student.ID,
	)

	return result.Error
}
