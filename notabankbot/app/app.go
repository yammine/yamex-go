package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/yammine/yamex-go/notabankbot/domain"
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
