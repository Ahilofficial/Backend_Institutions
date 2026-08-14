package model

import (
	"time"

	"gorm.io/gorm"
)

type Principal struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Gender      string         `gorm:"type:varchar(255)" json:"gender"`
	JoiningDate time.Time      `gorm:"type:date" json:"joining_date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"`
	IsActive    bool           `gorm:"default:true" json:"isactive"`

	InstitutionID uint `json:"institution_id"`
	UserID        uint `json:"user_id"`

	Institution *Institutions `gorm:"foreignKey:InstitutionID;references:ID" json:"institution,omitempty"`
	User        *User        `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
