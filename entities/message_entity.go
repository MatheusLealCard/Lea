package entities

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	SenderID uint `gorm:"not null;index"`
	Sender   User `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"json:"-"`

	RecieverID uint `gorm:"not null;index"`
	Reciever   User `gorm:"foreignKey:RecieverID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"json:"-"`

	Content string `gorm:"type:text;not null"`
}
