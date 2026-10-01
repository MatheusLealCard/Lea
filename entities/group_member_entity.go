package entities

import "gorm.io/gorm"

type GroupMember struct {
	gorm.Model
	GroupID uint   `gorm:"not null;uniqueIndex:idx_group_user"`
	Group   Group  `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	UserID  uint   `gorm:"not null;uniqueIndex:idx_group_user"`
	User    User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Role    string `gorm:"size:20;not null;default:'member'"`
}
