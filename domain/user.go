package domain

import "gorm.io/gorm"

type User struct {
	gorm.Model
	SlackID  string `gorm:"uniqueIndex"`
	Accounts []Account
}
