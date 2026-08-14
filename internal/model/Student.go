package model

import (
	"time"

	"gorm.io/gorm"
)

type Student struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Gender    string         `gorm:"type:varchar(255)" json:"gender"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
	IsActive  bool           `gorm:"default:true" json:"isactive"`

	FacultyID uint `json:"faculty_id"`
	UserID    uint `json:"user_id"`

	Faculty *Faculty `gorm:"foreignKey:FacultyID;references:ID" json:"faculty,omitempty"`
	User    *User    `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`

	Fees []Fees `gorm:"foreignKey:StudentID;references:ID" json:"fees,omitempty"`
}
