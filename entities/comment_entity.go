package entities

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	PostID  uint   `gorm:"not null;index"`
	Post    Post   `gorm:"foreignKey:PostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	UserID  uint   `gorm:"not null;index"`
	User    User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Content string `gorm:"type:text;not null"`
}
