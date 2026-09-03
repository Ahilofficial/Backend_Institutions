package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"strings"
	"github.com/jinzhu/copier"
)

type CreateFeesDTO struct {
	DepartmentID  uint    `json:"department_id"`
	Semester      uint    `json:"semester"`
	CollegeAmount float64 `json:"college_amount"`
	HostelAmount  float64 `json:"hostel_amount"`
	Amount        float64 `json:"amount"`
	PaymentMode   string  `json:"payment_mode"`
}

func (dto *CreateFeesDTO) Sanitize() {
	dto.PaymentMode = strings.TrimSpace(dto.PaymentMode)
	if dto.Semester == 0 {
		dto.Semester = 1
	}
	if dto.Amount == 0 && (dto.CollegeAmount > 0 || dto.HostelAmount > 0) {
		dto.Amount = dto.CollegeAmount + dto.HostelAmount
	}
}

func (dto *CreateFeesDTO) Validate() error {
	dto.Sanitize()
	if dto.DepartmentID == 0 {
		return errors.New("department_id is required")
	}
	if dto.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	return nil
}

type UpdateFeesDTO struct {
	CollegeAmount float64 `json:"college_amount"`
	HostelAmount  float64 `json:"hostel_amount"`
	Amount        float64 `json:"amount"`
	TotalAmount   float64 `json:"total_amount"`
	PaymentMode   string  `json:"payment_mode"`
}

func (dto *UpdateFeesDTO) Sanitize() {
	dto.PaymentMode = strings.TrimSpace(dto.PaymentMode)
	if dto.TotalAmount == 0 && dto.Amount > 0 {
		dto.TotalAmount = dto.Amount
	}
	if dto.TotalAmount == 0 && (dto.CollegeAmount > 0 || dto.HostelAmount > 0) {
		dto.TotalAmount = dto.CollegeAmount + dto.HostelAmount
	}
}

func (dto *UpdateFeesDTO) Validate() error {
	dto.Sanitize()
	if dto.TotalAmount <= 0 && dto.CollegeAmount <= 0 && dto.HostelAmount <= 0 && dto.PaymentMode == "" {
		return errors.New("no update fields provided")
	}
	return nil
}

type PaymentResponseDTO struct {
	ID         uint    `json:"id"`
	AmountPaid float64 `json:"amount_paid"`
}

type FeesResponseDTO struct {
	ID            uint                 `json:"id"`
	DepartmentID  uint                 `json:"department_id"`
	Semester      uint                 `json:"semester"`
	StudentID     *uint                `json:"student_id,omitempty"`
	CollegeAmount float64              `json:"college_amount"`
	HostelAmount  float64              `json:"hostel_amount"`
	TotalAmount   float64              `json:"amount"`
	TotalPaid     float64              `json:"total_paid"`
	PendingAmount float64              `json:"pending_amount"`
	PaymentMode   string               `json:"payment_mode,omitempty"`
	IsActive      bool                 `json:"is_active"`
	Payments      []PaymentResponseDTO `json:"payments,omitempty"`
}

func ToFeesResponseDTO(fees *model.Fees) FeesResponseDTO {
	var dto FeesResponseDTO
	copier.Copy(&dto, fees)
	return dto
}

func ToFeesResponseListDTO(fees []model.Fees) []FeesResponseDTO {
	list := make([]FeesResponseDTO, len(fees))
	for i, f := range fees {
		list[i] = ToFeesResponseDTO(&f)
	}
	return list
}

type CreatePaymentDTO struct {
	FeeID       uint    `json:"fee_id"`
	StudentID   uint    `json:"student_id,omitempty"`
	Semester    uint    `json:"semester,omitempty"`
	AmountPaid  float64 `json:"amount_paid"`
	PaymentMode string  `json:"payment_mode"`
}

func (dto *CreatePaymentDTO) Sanitize() {
	dto.PaymentMode = strings.TrimSpace(dto.PaymentMode)
}

func (dto *CreatePaymentDTO) Validate() error {
	dto.Sanitize()
	if dto.FeeID == 0 && dto.StudentID == 0 {
		return errors.New("fee_id or student_id is required")
	}
	if dto.AmountPaid <= 0 {
		return errors.New("amount_paid must be greater than zero")
	}
	if dto.PaymentMode == "" {
		return errors.New("payment_mode is required")
	}
	return nil
}

type StudentPaymentResponseDTO struct {
	ID          uint    `json:"id"`
	StudentID   uint    `json:"student_id"`
	PaymentID   uint    `json:"payment_id"`
	Semester    uint    `json:"semester"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}
