package adapter

import (
	"context"
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm/clause"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"

	"github.com/yammine/yamex-go/notabankbot/port"
)

type SlackCredentialPostgres struct {
	sync.RWMutex

	db    *gorm.DB
	cache map[string]string
}

func NewSlackCredentialPostgresRepository(dsn string) *SlackCredentialPostgres {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	return &SlackCredentialPostgres{
		db:    db,
		cache: make(map[string]string),
	}
}

func (s SlackCredentialPostgres) Migrate() error {
	return s.db.AutoMigrate(&SlackCredential{})
}

func (s SlackCredentialPostgres) SaveCredentials(ctx context.Context, workspaceID, token string) error {
	err := s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&SlackCredential{TeamID: workspaceID, Token: token}).Error
	if err != nil {
		return fmt.Errorf("insert credentials: %w", err)
	}
	// Update the cache
	s.Lock()
	defer s.Unlock()
	s.cache[workspaceID] = token

	return nil
}

func (s SlackCredentialPostgres) GetCredentials(ctx context.Context, workspaceID string) (string, error) {
	s.RLock()

	// Check cache first
	token, ok := s.cache[workspaceID]
	// Explicitly release read lock
	s.RUnlock()
	// That's a hit, return the token
	if ok {
		return token, nil
	}

	// Fallback to DB
	creds := &SlackCredential{}
	if err := s.db.Where("team_id = ?", workspaceID).First(creds).Error; err != nil {
		return "", fmt.Errorf("could not find credentials: %w", err)
	}
	// Populate cache for subsequent queries
	s.Lock()
	defer s.Unlock()
	s.cache[creds.TeamID] = creds.Token

	return creds.Token, nil
}

var _ port.SlackCredentialStore = (*SlackCredentialPostgres)(nil)
