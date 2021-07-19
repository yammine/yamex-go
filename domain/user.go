package domain

type User struct {
	ID      string
	SlackID string `fauna:"slack_user_id"`
}
