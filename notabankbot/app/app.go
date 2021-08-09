package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/yammine/yamex-go/notabankbot/domain"
)

const (
	// Top Level expressions

	CommandExpression    = "(?P<bot_id><@[A-Z0-9]{11}>)(?P<command>.+)(?P<recipient_id><@[A-Z0-9]{11}>)(?P<note>.*)"
	GetBalanceExpression = "(?P<bot_id><@[A-Z0-9]{11}>)[[:space:]](balance|my[[:space:]]balance)"

	// Sub-command expressions

	GrantCurrencyExpression = "grant[[:space:]]*(?P<currency>[$A-Za-z]+).*"

	// Command names

	GetBalanceCmd = "GetBalance"
	CommandCmd    = "Command"

	// Sub-command Names

	GrantCurrencyCmd = "GrantCurrency"

	// Responses

	GenericResponse      = "I don't understand what you're asking me :face_with_head_bandage:"
	GenericErrorResponse = "I seem to be experiencing an unexpected error :robot_face:"

	AlreadyGrantedCurrency = "Oops! Looks like you've already granted currency recently. Try again later :simple_smile:"

	// Capture Keys

	ckRecipientID = "recipient_id"
	ckCurrency    = "currency"
	ckNote        = "note"
	ckCommand     = "command"
)

type Application struct {
	repo Repository
}

func NewApplication(repo Repository) *Application {
	return &Application{
		repo: repo,
	}
}

type GrantInput struct {
	GranterID  string
	ReceiverID string
	Platform   string
	Currency   string
	Note       string
}

func (a Application) Grant(ctx context.Context, in *GrantInput) (*domain.Grant, error) {
	granter, err := a.repo.GetOrCreateUserBySlackID(ctx, in.GranterID)
	if err != nil {
		return nil, fmt.Errorf("fetching sender: %w", err)
	}
	recipient, err := a.repo.GetOrCreateUserBySlackID(ctx, in.ReceiverID)
	if err != nil {
		return nil, fmt.Errorf("fetching receiver: %w", err)
	}

	input := GrantCurrencyInput{
		Currency:   in.Currency,
		FromUserID: granter.ID,
		ToUserID:   recipient.ID,
		Note:       strings.TrimSpace(in.Note),
	}
	grant, err := a.repo.GrantCurrency(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("repo.GrantCurrency: %w", err)
	}
	return grant, nil
}

type GetBalanceInput struct {
	UserID string
}

func (a Application) GetBalance(ctx context.Context, in *GetBalanceInput) ([]*domain.Account, error) {
	user, err := a.repo.GetOrCreateUserBySlackID(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	return a.repo.GetAccountsForUser(ctx, user.ID)
}
