package model

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	DepartmentName string `gorm:"type:varchar(255);not null" json:"department_name"`
	CourseDuration uint   `gorm:"default:0" json:"course_duration"`
	IsActive       bool   `gorm:"default:true" json:"is_active"`

	InstitutionID uint          `gorm:"not null;index" json:"institution_id"`
	Institution   *Institutions `gorm:"foreignKey:InstitutionID;references:ID" json:"institution,omitempty"`

	Faculties []Faculty `gorm:"foreignKey:DepartmentID;references:ID" json:"faculties,omitempty"`

	Students []Student `gorm:"foreignKey:DepartmentID;references:ID" json:"students,omitempty"`

	SemesterFees []SemesterFee `gorm:"foreignKey:DepartmentID;references:ID" json:"semester_fees,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}