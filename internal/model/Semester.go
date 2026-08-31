package model

import (
	"time"

	"gorm.io/gorm"
)

type SemesterFee struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	DepartmentID uint `gorm:"not null;uniqueIndex:idx_department_semester" json:"department_id"`

	Semester uint `gorm:"not null;uniqueIndex:idx_department_semester" json:"semester"`

	CollegeAmount float64 `gorm:"type:decimal(10,2);default:0" json:"college_amount"`

	HostelAmount float64 `gorm:"type:decimal(10,2);default:0" json:"hostel_amount"`

	FeeAmount float64 `gorm:"type:decimal(10,2);default:0" json:"fee_amount"`

	IsActive bool `gorm:"default:true" json:"is_active"`

	Department *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}