package model

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID           uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	HostelAmount float64 `gorm:"type:decimal(10,2);default:0" json:"hostel_amount"`
	AmountPaid   float64 `gorm:"type:decimal(10,2);not null" json:"amount_paid"`
	PaymentMode  string  `gorm:"type:varchar(50)" json:"payment_mode"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`

	FeeID     uint  `gorm:"not null;index" json:"fee_id"`
	StudentID *uint `gorm:"default:null;index" json:"student_id,omitempty"`

	Fee     *Fees    `gorm:"foreignKey:FeeID;references:ID" json:"fee,omitempty"`
	Student *Student `gorm:"foreignKey:StudentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"student,omitempty"`
}
