package app

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

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
	receiver, err := a.repo.GetOrCreateUserBySlackID(ctx, in.ReceiverID)
	if err != nil {
		return nil, fmt.Errorf("fetching receiver: %w", err)
	}

	grant, err := a.repo.GrantCurrency(
		ctx,
		&GrantCurrencyInput{From: granter, To: receiver, Currency: in.Currency},
		func(ctx context.Context, gin *GrantCurrencyFuncIn) (*GrantCurrencyFuncOut, error) {
			if !gin.From.CanGrantCurrency() {
				return nil, domain.ErrAlreadyGranted
			}
			g := domain.NewGrant(gin.From, gin.To)
			m, _ := gin.ToAccount.Credit(decimal.New(1, 0), in.Note)

			return &GrantCurrencyFuncOut{
				Grant:    g,
				Movement: m,
			}, nil
		})

	if err != nil {
		return nil, fmt.Errorf("repo.GrantCurrency: %w", err)
	}

	return grant, nil
}

type TransferInput struct {
	SenderID   string
	ReceiverID string
	Platform   string
	Currency   string
	Amount     decimal.Decimal
	Note       string
}

func (a Application) Transfer(ctx context.Context, input *TransferInput) error {
	sender, err := a.repo.GetOrCreateUserBySlackID(ctx, input.SenderID)
	if err != nil {
		return fmt.Errorf("fetching sender: %w", err)
	}
	receiver, err := a.repo.GetOrCreateUserBySlackID(ctx, input.ReceiverID)
	if err != nil {
		return fmt.Errorf("fetching receiver: %w", err)
	}

	return a.repo.SendCurrency(ctx,
		&SendCurrencyInput{
			From:     sender,
			To:       receiver,
			Currency: input.Currency,
		}, func(ctx context.Context, in *SendCurrencyFuncIn) (*SendCurrencyFuncOut, error) {
			// debit the sender
			debit, err := in.FromAccount.Debit(input.Amount, input.Note)
			if err != nil {
				return nil, err
			}
			// credit the receiver
			credit, _ := in.ToAccount.Credit(input.Amount, input.Note)

			return &SendCurrencyFuncOut{
				SendingMovement:   debit,
				ReceivingMovement: credit,
			}, nil
		})
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

func (a Application) SaveFeedback(ctx context.Context, slackUserID, feedback string) error {
	user, err := a.repo.GetOrCreateUserBySlackID(ctx, slackUserID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	return a.repo.SaveFeedback(ctx, user, feedback)
}
