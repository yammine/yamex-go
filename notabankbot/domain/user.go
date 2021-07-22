package domain

import "gorm.io/gorm"

type User struct {
	gorm.Model
	SlackID string `gorm:"uniqueIndex"`

	Accounts       []Account
	GrantsGiven    []Grant `gorm:"foreignKey:FromUserID"`
	GrantsReceived []Grant `gorm:"foreignKey:ToUserID"`
}
