package entities

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	SenderID uint `gorm:"not null;index"`
	Sender   User `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	RecieverID uint `gorm:"not null;index"`
	Reciever   User `gorm:"foreignKey:RecieverID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	Content string `gorm:"type:text;not null"`

	// post compartilhado (opcional). Se o post for apagado, a mensagem continua
	PostID *uint `gorm:"index"`
	Post   *Post `gorm:"foreignKey:PostID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}
