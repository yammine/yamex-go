package app

import "github.com/yammine/yamex-go/domain"

type Repository interface {
	GetOrCreateUser(slackUserId string) (*domain.User, error)
}
