package model

import (
	"time"

	"gorm.io/gorm"
)

type DepartmentPayment struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DepartmentID  uint           `gorm:"not null;index" json:"department_id"`
	Semester      uint           `gorm:"not null;index" json:"semester"`
	CollegeAmount float64        `gorm:"type:decimal(10,2);not null" json:"college_amount"`
	HostelAmount  float64        `gorm:"type:decimal(10,2);default:0" json:"hostel_amount"`
	TotalAmount   float64        `gorm:"type:decimal(10,2);default:0" json:"total_amount"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Department *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`
}
