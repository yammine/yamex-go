package adapter

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yammine/yamex-go/app"
	"github.com/yammine/yamex-go/domain"
)

func NewPostgresRepository(dsn string) *PostgresRepository {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	return &PostgresRepository{DB: db}
}

type PostgresRepository struct {
	DB *gorm.DB
}

func (p PostgresRepository) Migrate() error {
	return p.DB.AutoMigrate(&domain.User{}, &domain.Account{})
}

func (p PostgresRepository) GetOrCreateUserBySlackID(ctx context.Context, slackUserId string) (*domain.User, error) {
	user := domain.User{
		SlackID: slackUserId,
	}
	account := domain.Account{
		Currency: "$yam",
	}

	tx := p.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, fmt.Errorf("starting txn: %w", err)
	}

	// Insert user
	insertUserResult := tx.FirstOrCreate(&user, user)
	if err := insertUserResult.Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("insert user: %w", err)
	}

	account.UserID = user.ID
	insertAccountResult := tx.FirstOrCreate(&account, account)
	if err := insertAccountResult.Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("insert account: %w", err)
	}

	user.Accounts = append(user.Accounts, account)

	return &user, tx.Commit().Error
}

var _ app.Repository = (*PostgresRepository)(nil)
