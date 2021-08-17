package port

import "context"

type SlackCredentialStore interface {
	SaveCredentials(ctx context.Context, workspaceID, token string) error
	GetCredentials(ctx context.Context, workspaceID string) (string, error)
}
