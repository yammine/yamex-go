package domain

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model

	UserID   uint            `gorm:"index:idx_accounts_user_id_currency,unique"`
	Currency string          `gorm:"index:idx_accounts_user_id_currency,unique"`
	Balance  decimal.Decimal `gorm:"type:decimal(20,8);"`

	Movements []Movement
}

func (a Account) UpdateBalance(m *Movement) {
	a.Balance = a.Balance.Add(m.Amount)
}
