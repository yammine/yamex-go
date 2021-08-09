package domain

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Movement struct {
	gorm.Model
	AccountID uint
	Amount    decimal.Decimal `gorm:"type:decimal(20,8)"`
	Reason    string
}

func NewMovement(account *Account, amount decimal.Decimal, reason string) *Movement {
	return &Movement{
		AccountID: account.ID,
		Amount:    amount,
		Reason:    reason,
	}
}
