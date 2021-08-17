package adapter

// TODO: Use adapter defined models for marshalling/unmarshalling DB values.

//
//import (
//	"time"
//
//	"github.com/shopspring/decimal"
//	"gorm.io/gorm"
//)
//
//type Account struct {
//	gorm.Model
//
//	UserID   uint            `gorm:"index:idx_accounts_user_id_currency,unique"`
//	Currency string          `gorm:"index:idx_accounts_user_id_currency,unique"`
//	Balance  decimal.Decimal `gorm:"type:decimal(20,8);"`
//
//	Movements []*Movement
//}
//type Grant struct {
//	ID        uint      `gorm:"primarykey"`
//	CreatedAt time.Time `gorm:"index"`
//
//	FromUserID uint
//	FromUser   *User
//	ToUserID   uint
//	ToUser     *User
//	MovementID uint `gorm:"index"`
//	Movement   *Movement
//}
//
//type Movement struct {
//	gorm.Model
//	AccountID uint
//	Amount    decimal.Decimal `gorm:"type:decimal(20,8)"`
//	Reason    string
//}
//
//type User struct {
//	gorm.Model
//	SlackID string `gorm:"uniqueIndex"`
//	Admin   bool
//
//	Accounts            []*Account
//	RecentlyGivenGrants []*Grant `gorm:"foreignKey:FromUserID"`
//	GrantsGiven         []*Grant `gorm:"foreignKey:FromUserID"`
//	GrantsReceived      []*Grant `gorm:"foreignKey:ToUserID"`
//}