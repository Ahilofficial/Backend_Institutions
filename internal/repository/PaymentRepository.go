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
	err := r.db.Where("department_id = ? AND semester = ? AND deleted_at IS NULL", deptID, semester).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) GetDepartmentPaymentByID(id uint) (*model.DepartmentPayment, error) {
	var p model.DepartmentPayment
	err := r.db.Preload("Department").Where("id = ? AND deleted_at IS NULL", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) GetDepartmentPaymentsByDeptID(deptID uint) ([]model.DepartmentPayment, error) {
	var list []model.DepartmentPayment
	err := r.db.Where("department_id = ? AND deleted_at IS NULL", deptID).Order("semester ASC").Find(&list).Error
	return list, err
}

func (r *PaymentRepository) UpdateDepartmentPayment(p *model.DepartmentPayment) error {
	return r.db.Model(&model.DepartmentPayment{}).
		Where("id = ? AND deleted_at IS NULL", p.ID).
		Updates(map[string]interface{}{
			"college_amount": p.CollegeAmount,
			"hostel_amount":  p.HostelAmount,
			"total_amount":   p.TotalAmount,
			"updated_at":     time.Now(),
		}).Error
}

func (r *PaymentRepository) DeleteDepartmentPayment(id uint) error {
	res := r.db.Where("id = ? AND deleted_at IS NULL", id).Delete(&model.DepartmentPayment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("payment record not found or already deleted")
	}
	return nil
}

func (r *PaymentRepository) CreateStudentPayment(p *model.StudentPayment) error {
	return r.db.Create(p).Error
}

func (r *PaymentRepository) GetStudentPaymentsByStudentID(studentID uint) ([]model.StudentPayment, error) {
	var payments []model.StudentPayment
	err := r.db.Where("student_id = ? AND deleted_at IS NULL", studentID).Order("created_at DESC").Find(&payments).Error
	return payments, err
}

func (r *PaymentRepository) GetStudentByID(studentID uint) (*model.Student, error) {
	var s model.Student
	err := r.db.Where("id = ? AND deleted_at IS NULL", studentID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PaymentRepository) UpdateStudentPaymentStatus(student *model.Student) error {
	return r.db.Model(&model.Student{}).
		Where("id = ? AND deleted_at IS NULL", student.ID).
		Updates(map[string]interface{}{
			"paid_amount": student.PaidAmount,
			"pending":     student.Pending,
			"updated_at":  time.Now(),
		}).Error
}
