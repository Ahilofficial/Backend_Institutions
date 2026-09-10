package model

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);unique;not null" json:"name"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`

	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}
