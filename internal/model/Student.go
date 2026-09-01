package model

import (
	"time"

	"gorm.io/gorm"
)

type Student struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Gender    string         `gorm:"type:varchar(255)" json:"gender"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Hosteller   bool `gorm:"default:false" json:"hosteller"`
	Scholarship bool `gorm:"default:false" json:"scholarship"`
	MQ          bool `gorm:"default:false" json:"mq"`

	BaseAmount float64 `gorm:"type:decimal(10,2);default:0" json:"base_amount"`
	FeeAmount  float64 `gorm:"type:decimal(10,2);default:0" json:"fee_amount"`

	Semester uint `gorm:"not null;default:1" json:"semester"`

	Pending bool `gorm:"default:true" json:"pending"`

	IsProfileVerified bool `gorm:"default:false" json:"is_verified"`

	UserID       uint `gorm:"not null;index" json:"user_id"`
	FacultyID    uint `gorm:"not null;index" json:"faculty_id"`
	DepartmentID uint `gorm:"not null;index" json:"department_id"`

	User       *User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Faculty    *Faculty    `gorm:"foreignKey:FacultyID;references:ID" json:"faculty,omitempty"`
	Department *Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department,omitempty"`

	Fees            []Fees           `gorm:"foreignKey:StudentID;references:ID" json:"fees,omitempty"`
	StudentPayments []StudentPayment `gorm:"foreignKey:StudentID;references:ID" json:"student_payments,omitempty"`
}
