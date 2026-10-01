package entities

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name        string `gorm:"size:100;not null;uniqueIndex"`
	Description string `gorm:"size:500;not null;default:''"`
	OwnerID     uint   `gorm:"not null;index"`
	Owner       User   `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
