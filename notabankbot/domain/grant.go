package domain

import (
	"time"

	"github.com/yammine/yamex-go"
)

const ErrAlreadyGranted = yamex.Sentinel("already granted currency")

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

func NewGrant(from, to *User) *Grant {
	return &Grant{
		FromUserID: from.ID,
		ToUserID:   to.ID,
	}
}

func TimeBetweenGrants() time.Duration {
	return 10 * time.Second
}
