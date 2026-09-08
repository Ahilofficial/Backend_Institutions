package model

import (
	"time"

	"gorm.io/gorm"
)

type StudentPayment struct {
	ID                  uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID           uint           `gorm:"not null;index" json:"student_id"`
	DepartmentPaymentID uint           `gorm:"not null;index" json:"fee_id"`
	AmountPaid          float64        `gorm:"type:decimal(10,2);not null" json:"amount_paid"`
	PaymentMode         string         `gorm:"type:varchar(50);not null" json:"payment_mode"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	Student           *Student           `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"`
	DepartmentPayment *DepartmentPayment `gorm:"foreignKey:DepartmentPaymentID;references:ID" json:"department_payment,omitempty"`
}
