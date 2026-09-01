package model

import (
	"time"

	"gorm.io/gorm"
)

type StudentPayment struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID   uint           `gorm:"not null" json:"student_id"`
	PaymentID   uint           `gorm:"not null" json:"payment_id"`
	Semester    uint           `gorm:"not null;default:1" json:"semester"`
	TotalAmount float64        `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"`

	Student *Student `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"`
}
