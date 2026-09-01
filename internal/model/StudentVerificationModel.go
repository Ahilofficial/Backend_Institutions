package model

import (
	"time"

	"gorm.io/gorm"
)

type StudentVerificationAccess struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	StudentID uint `gorm:"uniqueIndex:idx_student_faculty" json:"student_id"`
	FacultyID uint `gorm:"uniqueIndex:idx_student_faculty" json:"faculty_id"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `json:"-"`
}
