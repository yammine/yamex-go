package app

import (
	"context"

	"github.com/yammine/yamex-go"

	"github.com/yammine/yamex-go/notabankbot/domain"
)

const (
	ErrCannotFindOrCreateUser = yamex.Sentinel("cannot find or create user")
)

type GrantFunc = func(ctx context.Context, in *GrantCurrencyFuncIn) (*GrantCurrencyFuncOut, error)
type SendFunc = func(ctx context.Context, in *SendCurrencyFuncIn) (*SendCurrencyFuncOut, error)

type Repository interface {
	GrantCurrency(ctx context.Context, in *GrantCurrencyInput, grantFn GrantFunc) (*domain.Grant, error)
	SendCurrency(ctx context.Context, in *SendCurrencyInput, sendFn SendFunc) error

	GetOrCreateUserBySlackID(ctx context.Context, slackUserId string) (*domain.User, error)
	GetAccountsForUser(ctx context.Context, id uint) ([]*domain.Account, error)

	SaveFeedback(ctx context.Context, user *domain.User, feedback string) error
}

// GrantCurrency

type GrantCurrencyInput struct {
	From     *domain.User
	To       *domain.User
	Currency string
}

type GrantCurrencyFuncIn struct {
	From      *domain.User
	To        *domain.User
	ToAccount *domain.Account
}

type GrantCurrencyFuncOut struct {
	Movement *domain.Movement
	Grant    *domain.Grant
}

// SendCurrency

type SendCurrencyInput struct {
	From     *domain.User
	To       *domain.User
	Currency string
}

type SendCurrencyFuncIn struct {
	FromAccount *domain.Account
	ToAccount   *domain.Account
}

type SendCurrencyFuncOut struct {
	SendingMovement   *domain.Movement
	ReceivingMovement *domain.Movement
}
