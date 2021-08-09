package adapter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yammine/yamex-go/notabankbot/app"
	"github.com/yammine/yamex-go/notabankbot/domain"

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

func (p PostgresRepository) GrantCurrency(ctx context.Context, input *app.GrantCurrencyInput, grantFn app.GrantFunc) (*domain.Grant, error) {
	var grant *domain.Grant
	err := p.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		from, txErr := getUserExclusive(tx, input.From.ID)
		if txErr != nil {
			return fmt.Errorf("get sender user exclusive: %w", txErr)
		}
		account, txErr := getAccountExclusive(tx, input.To.ID, input.Currency)
		if txErr != nil {
			return fmt.Errorf("get receiver account exclusive: %w", txErr)
		}

		out, txErr := grantFn(ctx, &app.GrantCurrencyFuncIn{
			From:      from,
			To:        input.To,
			ToAccount: account,
		})
		if txErr != nil {
			return fmt.Errorf("business logic error: %w", txErr)
		}

		// Save the updated account balance
		if saveAccountErr := tx.Save(out.UpdatedAccount).Error; saveAccountErr != nil {
			return fmt.Errorf("saving updated account: %w", saveAccountErr)
		}

		if insertMovementErr := tx.Create(out.Movement).Error; insertMovementErr != nil {
			return fmt.Errorf("inserting movement: %w", insertMovementErr)
		}

		// Associate the newly inserted movement with the grant.
		out.Grant.MovementID = out.Movement.ID
		if insertGrantErr := tx.Create(out.Grant).Error; insertGrantErr != nil {
			return fmt.Errorf("inserting grant: %w", insertGrantErr)
		}

		return nil
	})

	return grant, err
}

var _ app.Repository = (*PostgresRepository)(nil)

func getUserExclusive(tx *gorm.DB, id uint) (*domain.User, error) {
	var user domain.User

	err := tx.
		Preload("RecentlyGivenGrants", "created_at >= ?", time.Now().Add(-domain.TimeBetweenGrants())).
		Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, id).
		Error
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	return &user, nil
}

func getAccountExclusive(tx *gorm.DB, id uint, currency string) (*domain.Account, error) {
	var account *domain.Account
	txErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).FirstOrCreate(&account, domain.Account{UserID: id, Currency: currency}).Error
	if txErr != nil {
		return nil, fmt.Errorf("get account: %w", txErr)
	}

	return account, nil
}
