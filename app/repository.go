package app

import "github.com/yammine/yamex-go/domain"

type Repository interface {
	GetOrCreateUser(id string) (*domain.User, error)
}
