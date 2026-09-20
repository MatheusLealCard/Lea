package entities

import (
	"gorm.io/gorm"
)

type Follower struct {
	gorm.Model
	UserID uint `gorm:"not null;uniqueIndex:idx_user_folower"`
	User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	FolowerID uint `gorm:"not null;uniqueIndex:idx_user_folower"`
	Folower   User `gorm:"foreignKey:FolowerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
