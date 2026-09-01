package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(255)" json:"name"`
	Email    string `gorm:"type:varchar(255)" json:"email"`
	Phone    string `gorm:"type:varchar(255)" json:"phone"`
	Password string `gorm:"type:varchar(255)" json:"-"`

	IsActive   bool `gorm:"type:boolean;default:true" json:"is_active"`
	IsVerified bool `json:"is_verified"`

	VerificationToken   string    `json:"-"`
	TokenExpiresAt      time.Time `json:"-"`
	ResetPasswordToken  string    `json:"-"`
	ResetTokenExpiresAt time.Time `json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`

	StudentID uint `json:"student_id"`
	FacultyID uint `json:"faculty_id"`

	Student *Student `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"`
	Faculty *Faculty `gorm:"foreignKey:FacultyID;references:ID" json:"faculty,omitempty"`

	Roles []Role `gorm:"many2many:user_roles" json:"roles,omitempty"`
}
