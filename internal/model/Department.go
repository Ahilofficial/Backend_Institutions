package model

import (
	"time"

	"gorm.io/gorm"
)
type Department struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DepartmentName string         `gorm:"type:varchar(255)" json:"department_name"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-"`
	IsActive       bool           `gorm:"default:true" json:"isactive"`

	PaymentID     uint             `json:"payment_id"`

	InstitutionID uint          `json:"institution_id"`
	Institution   *Institutions `gorm:"foreignKey:InstitutionID;references:ID" json:"institution,omitempty"`

	Faculties []Faculty `gorm:"foreignKey:DepartmentID;references:ID" json:"faculties,omitempty"`
}
