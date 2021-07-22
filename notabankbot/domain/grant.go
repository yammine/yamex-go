package domain

import (
	"time"

	"github.com/yammine/yamex-go"
)

const ErrAlreadyGrantedWithinThreeDays = yamex.Sentinel("already granted currency in the past 3 days")

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
