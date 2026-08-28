package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"github.com/jinzhu/copier"
	"strings"
)

type CreateFeesDTO struct {
	PaymentMode              string  `json:"payment_mode"`
	TotalAmount              float64 `json:"total_amount"`
	Amount                   float64 `json:"amount"`
	StudentID                uint    `json:"student_id"`
}

func (dto *CreateFeesDTO) Sanitize() {
	dto.PaymentMode = strings.TrimSpace(strings.ToLower(dto.PaymentMode))
	if dto.TotalAmount == 0 && dto.Amount > 0 {
		dto.TotalAmount = dto.Amount
	}
}

func (dto *CreateFeesDTO) Validate() error {
	if dto.TotalAmount == 0 && dto.Amount > 0 {
		dto.TotalAmount = dto.Amount
	}
	if dto.PaymentMode == "" {
		return errors.New("payment mode is required")
	}
	if dto.TotalAmount <= 0 {
		return errors.New("amount is required and must be greater than 0")
	}
	if dto.StudentID == 0 {
		return errors.New("student id is required")
	}
	return nil
}

type UpdateFeesDTO struct {
	PaymentMode string  `json:"payment_mode"`
	Amount      float64 `json:"amount"`
	TotalAmount float64 `json:"total_amount"`
}

func (dto *UpdateFeesDTO) Sanitize() {
	dto.PaymentMode = strings.TrimSpace(strings.ToLower(dto.PaymentMode))
	if dto.Amount == 0 && dto.TotalAmount > 0 {
		dto.Amount = dto.TotalAmount
	}
}

func (dto *UpdateFeesDTO) Validate() error {
	dto.Sanitize()

	if dto.PaymentMode == "" {
		return errors.New("payment mode is required")
	}
	if dto.Amount <= 0 {
		return errors.New("amount is required and must be greater than 0")
	}
	return nil
}

type PaymentResponseDTO struct {
	ID          uint    `json:"id"`
	Month       string  `json:"month"`
	AmountPaid  float64 `json:"amount_paid"`
	PaymentMode string  `json:"payment_mode"`
}
type FeesResponseDTO struct {
	ID            uint                 `json:"id"`
	TotalAmount   float64              `json:"total_amount"`
	TotalPaid     float64              `json:"total_paid"`
	PendingAmount float64              `json:"pending_amount"`
	StudentID     uint                 `json:"student_id"`
	IsActive      bool                 `json:"is_active"`
	Payments      []PaymentResponseDTO `json:"payments"`
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
	Month       string  `json:"month"`
	AmountPaid  float64 `json:"amount_paid"`
	PaymentMode string  `json:"payment_mode"`
	FeeID       uint    `json:"fee_id"`
}

func (dto *CreatePaymentDTO) Sanitize() {
	dto.Month = strings.TrimSpace(dto.Month)
	dto.PaymentMode = strings.TrimSpace(strings.ToLower(dto.PaymentMode))
}

func (dto *CreatePaymentDTO) Validate() error {
	if dto.Month == "" {
		return errors.New("month is required")
	}

	if dto.PaymentMode == "" {
		return errors.New("payment mode is required")
	}

	if dto.AmountPaid <= 0 {
		return errors.New("amount paid must be greater than zero")
	}

	if dto.FeeID == 0 {
		return errors.New("fee id is required")
	}

	return nil
}
