package entities

import (
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	RoleName string `gorm:"size:20;not null;unique"`
}
