package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/yammine/yamex-go"
	"gorm.io/gorm"
)

const (
	ErrAmountCannotBeNegative yamex.Sentinel = "amount cannot be negative"
	ErrInsufficientBalance    yamex.Sentinel = "insufficient balance"
)

type Account struct {
	gorm.Model

	UserID   uint            `gorm:"index:idx_accounts_user_id_currency,unique"`
	Currency string          `gorm:"index:idx_accounts_user_id_currency,unique"`
	Balance  decimal.Decimal `gorm:"type:decimal(20,8);"`

	Movements []Movement
}

func (a *Account) ApplyNewMovement(m *Movement) {
	a.Balance = a.Balance.Add(m.Amount)
}

func (a *Account) Debit(amount decimal.Decimal, reason string) (*Movement, error) {
	if amount.IsNegative() {
		return nil, ErrAmountCannotBeNegative
	}

	fmt.Printf("Account %+v\n", a)
	newBalance := a.Balance.Sub(amount)
	fmt.Printf("New balance: %s", newBalance.String())
	if newBalance.IsNegative() {
		return nil, ErrInsufficientBalance
	}

	a.Balance = newBalance

	movement := NewMovement(a, amount.Neg(), fmt.Sprintf("out: %s", reason))
	return movement, nil
}

func (a *Account) Credit(amount decimal.Decimal, reason string) (*Movement, error) {
	// TODO: Figure out if there's a reason this could error.
	newBalance := a.Balance.Add(amount)
	a.Balance = newBalance

	movement := NewMovement(a, amount, reason)
	return movement, nil
}
