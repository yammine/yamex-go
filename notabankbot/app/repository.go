package app

import (
	"context"

	"github.com/yammine/yamex-go"

	"github.com/yammine/yamex-go/notabankbot/domain"
)

const (
	ErrCannotFindOrCreateUser = yamex.Sentinel("cannot find or create user")
)

type Repository interface {
	GetOrCreateUserBySlackID(ctx context.Context, slackUserId string) (*domain.User, error)
	GrantCurrency(ctx context.Context, input GrantCurrencyInput) (*domain.Grant, error)
	GetAccountsForUser(ctx context.Context, id uint) ([]*domain.Account, error)
}

type GrantCurrencyInput struct {
	Currency   string
	FromUserID uint
	ToUserID   uint
	Note       string
}
