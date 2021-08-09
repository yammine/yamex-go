package domain

import "gorm.io/gorm"

type User struct {
	gorm.Model
	SlackID string `gorm:"uniqueIndex"`
	Admin   bool

	Accounts            []Account
	RecentlyGivenGrants []*Grant `gorm:"foreignKey:FromUserID"`
	GrantsGiven         []Grant  `gorm:"foreignKey:FromUserID"`
	GrantsReceived      []Grant  `gorm:"foreignKey:ToUserID"`
}

// CanGrantCurrency assumes RecentlyGivenGrants has been loaded.
func (u User) CanGrantCurrency() bool {
	// Admins can always grant currency
	if u.Admin {
		return true
	}

	// Non-admins can only grant if they have no recent grants
	return len(u.RecentlyGivenGrants) < 1
}
