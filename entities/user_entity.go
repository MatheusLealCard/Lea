package entities

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `gorm:"size:100;not null"`
	Email    string `gorm:"size:100;not null;unique"`
	Password string `gorm:"size:255;not null" json:"-"`
	Avatar   string `gorm:"size:255;not null;default:''"`
	RoleID   uint   `gorm:"not null;default:1"`
	Role     Role   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}
