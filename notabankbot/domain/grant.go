package domain

import (
	"errors"
	"time"
)

type Grant struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`

	FromUserID uint
	FromUser   User
	ToUserID   uint
	ToUser     User
	MovementID uint `gorm:"index"`
	Movement   Movement
}

func ErrAlreadyGrantedWithinThreeDays() error {
	return errors.New("already granted currency in the past 3 days")
}
