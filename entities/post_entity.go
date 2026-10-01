package entities

import (
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title      string   `gorm:"size:150;not null"`
	Content    string   `gorm:"type:text;not null"`
	Photo      string   `gorm:"size:255;not null;default:''"`
	UserID     uint     `gorm:"not null"`
	User       User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CategoryID uint     `gorm:"not null"`
	Category   Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
