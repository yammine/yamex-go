package app

import (
	"context"

	"github.com/yammine/yamex-go/domain"
)

type Repository interface {
	GetOrCreateUserBySlackID(ctx context.Context, slackUserId string) (*domain.User, error)
}
