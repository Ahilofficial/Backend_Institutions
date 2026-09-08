package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"strings"

	"github.com/jinzhu/copier"
)

type CreateDepartmentPaymentDTO struct {
	CollegeAmount float64 `json:"college_amount"`
	HostelAmount  float64 `json:"hostel_amount"`
	DepartmentID  uint    `json:"department_id"`
	Semester      uint    `json:"semester"`
}

func (dto *CreateDepartmentPaymentDTO) Validate() error {
	if dto.DepartmentID == 0 {
		return errors.New("department_id is required")
	}
	if dto.Semester == 0 {
		return errors.New("semester is required")
	}
	if dto.CollegeAmount < 0 {
		return errors.New("college_amount cannot be negative")
	}
	if dto.HostelAmount < 0 {
		return errors.New("hostel_amount cannot be negative")
	}
	if dto.CollegeAmount == 0 && dto.HostelAmount == 0 {
		return errors.New("at least one of college_amount or hostel_amount must be greater than 0")
	}
	return nil
}

type UpdateDepartmentPaymentDTO struct {
	CollegeAmount float64 `json:"college_amount"`
	HostelAmount  float64 `json:"hostel_amount"`
}

func (dto *UpdateDepartmentPaymentDTO) Validate() error {
	if dto.CollegeAmount < 0 || dto.HostelAmount < 0 {
		return errors.New("amount cannot be negative")
	}
	return nil
}

type DepartmentPaymentResponseDTO struct {
	ID            uint    `json:"id"`
	DepartmentID  uint    `json:"department_id"`
	Semester      uint    `json:"semester"`
	CollegeAmount float64 `json:"college_amount"`
	HostelAmount  float64 `json:"hostel_amount"`
	TotalAmount   float64 `json:"total_amount"`
}

func ToDepartmentPaymentResponseDTO(p *model.DepartmentPayment) DepartmentPaymentResponseDTO {
	var resp DepartmentPaymentResponseDTO
	copier.Copy(&resp, p)
	return resp
}

func ToDepartmentPaymentResponseListDTO(list []model.DepartmentPayment) []DepartmentPaymentResponseDTO {
	res := make([]DepartmentPaymentResponseDTO, len(list))
	for i, item := range list {
		res[i] = ToDepartmentPaymentResponseDTO(&item)
	}
	return res
}

type CreateStudentPaymentDTO struct {
	AmountPaid  float64 `json:"amount_paid"`
	PaymentMode string  `json:"payment_mode"`
	StudentID   uint    `json:"student_id"`
	FeeID       uint    `json:"fee_id"`
}

func (dto *CreateStudentPaymentDTO) Sanitize() {
	dto.PaymentMode = strings.TrimSpace(dto.PaymentMode)
}

func (dto *CreateStudentPaymentDTO) Validate() error {
	dto.Sanitize()
	if dto.AmountPaid <= 0 {
		return errors.New("amount_paid must be greater than zero")
	}
	if dto.PaymentMode == "" {
		return errors.New("payment_mode is required")
	}
	return nil
}

type StudentPaymentResponseDTO struct {
	ID                  uint    `json:"id"`
	StudentID           uint    `json:"student_id"`
	DepartmentPaymentID uint    `json:"fee_id"`
	AmountPaid          float64 `json:"amount_paid"`
	PaymentMode         string  `json:"payment_mode"`
	StudentFeeAmount    float64 `json:"student_fee_amount"`
	StudentPaidAmount   float64 `json:"student_paid_amount"`
	Pending             bool    `json:"pending"`
}

func ToStudentPaymentResponseDTO(p *model.StudentPayment, student *model.Student) StudentPaymentResponseDTO {
	resp := StudentPaymentResponseDTO{
		ID:                  p.ID,
		StudentID:           p.StudentID,
		DepartmentPaymentID: p.DepartmentPaymentID,
		AmountPaid:          p.AmountPaid,
		PaymentMode:         p.PaymentMode,
	}
	if student != nil {
		resp.StudentFeeAmount = student.FeeAmount
		resp.StudentPaidAmount = student.PaidAmount
		resp.Pending = student.Pending
	}
	return resp
}
