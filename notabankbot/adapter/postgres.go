package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/yammine/yamex-go/notabankbot/app"
	"github.com/yammine/yamex-go/notabankbot/domain"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	return p.DB.AutoMigrate(&domain.User{}, &domain.Account{}, &domain.Movement{}, &domain.Grant{})
}

func (p PostgresRepository) GetAccountsForUser(ctx context.Context, id uint) ([]*domain.Account, error) {
	var accounts []*domain.Account

	if err := p.DB.WithContext(ctx).Find(&accounts, domain.Account{UserID: id}).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

func (p PostgresRepository) GetOrCreateUserBySlackID(ctx context.Context, slackUserId string) (*domain.User, error) {
	// Guard against bunk input, should probably move this up to the port/app
	if slackUserId == "" {
		return nil, app.ErrCannotFindOrCreateUser
	}
	user := domain.User{
		SlackID: slackUserId,
	}

	tx := p.DB.WithContext(ctx).FirstOrCreate(&user, user)

	return &user, tx.Error
}

func (p PostgresRepository) GrantCurrency(ctx context.Context, input app.GrantCurrencyInput) (*domain.Grant, error) {
	var grant domain.Grant
	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		fromUser, err := getUserExclusive(tx, input.FromUserID)
		if err != nil {
			return fmt.Errorf("from user: %w", err)
		}
		// Check to see if the granting user has granted in the past duration
		var formerGrant domain.Grant
		result := tx.Where(
			"from_user_id = @fromUserId AND created_at >= @threshold",
			sql.Named("fromUserId", fromUser.ID),
			sql.Named("threshold", time.Now().Add(-domain.TimeBetweenGrants())),
		).Find(&formerGrant)
		if result.RowsAffected > 0 {
			return domain.ErrAlreadyGranted
		}

		// Row exclusive lock to ensure we're the only one interacting with this user for the duration of tx
		toUser, err := getUserExclusive(tx, input.ToUserID)
		if err != nil {
			return fmt.Errorf("to user: %w", err)
		}

		// Fetch or create an account for this currency
		account := domain.Account{Currency: input.Currency, UserID: input.ToUserID}
		if err := tx.FirstOrCreate(&account, account).Error; err != nil {
			return fmt.Errorf("fetching account: %w", err)
		}

		// Insert Movement
		movement := domain.Movement{
			AccountID: account.ID,
			Amount:    decimal.NewFromFloat(1.0),
			Reason:    input.Note,
		}
		if err := tx.Create(&movement).Error; err != nil {
			return fmt.Errorf("insert movement: %w", err)
		}

		// Persist new Account.Balance
		tx.Model(&account).Select("balance").Updates(map[string]interface{}{"balance": account.Balance.Add(movement.Amount)})

		grant.FromUserID = input.FromUserID
		grant.ToUserID = toUser.ID
		grant.MovementID = movement.ID
		if err := tx.Create(&grant).Error; err != nil {
			return fmt.Errorf("insert grant: %w", err)
		}

		return nil
	})

	return &grant, err
}

var _ app.Repository = (*PostgresRepository)(nil)

func getUserExclusive(tx *gorm.DB, id uint) (*domain.User, error) {
	var user domain.User

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, id).Error
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	return &user, nil
}
