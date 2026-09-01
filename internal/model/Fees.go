package model

import (
	"time"

	"gorm.io/gorm"
)

type Fees struct {
	ID           uint  `gorm:"primaryKey;autoIncrement" json:"id"`
	DepartmentID uint  `gorm:"not null;index" json:"department_id"`
	Semester     uint  `gorm:"not null;default:1" json:"semester"`
	StudentID    *uint `gorm:"index" json:"student_id,omitempty"`

	CollegeAmount float64 `gorm:"type:decimal(10,2);default:0" json:"college_amount"`
	HostelAmount  float64 `gorm:"type:decimal(10,2);default:0" json:"hostel_amount"`
	TotalAmount   float64 `gorm:"type:decimal(10,2);default:0" json:"amount"`
	TotalPaid     float64 `gorm:"type:decimal(10,2);default:0" json:"total_paid"`
	PendingAmount float64 `gorm:"type:decimal(10,2);default:0" json:"pending_amount"`
	PaymentMode   string  `gorm:"type:varchar(255)" json:"payment_mode"`
	IsActive      bool    `gorm:"default:true" json:"isactive"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`

	Student    *Student    `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"`
	Department *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`
	Payments   []Payment   `gorm:"foreignKey:FeeID;references:ID" json:"payments,omitempty"`
}
