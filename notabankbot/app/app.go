package app

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	CommandExpression    = "(?P<bot_id><@[A-Z0-9]{11}>)(?P<command>.+)(?P<recipient_id><@[A-Z0-9]{11}>)(?P<note>.*)"
	GetBalanceExpression = "(?P<bot_id><@[A-Z0-9]{11}>)[[:space:]](balance|my[[:space:]]balance)"

	GenericResponse      = "I don't understand what you're asking me :face_with_head_bandage:"
	GenericErrorResponse = "I seem to be experiencing an unexpected error :robot_face:"
)

type Application struct {
	expressions           []*regexp.Regexp
	subCommandExpressions []*regexp.Regexp
	repo                  Repository
}

func NewApplication(repo Repository) *Application {
	top := []*regexp.Regexp{
		regexp.MustCompile(CommandExpression),
		regexp.MustCompile(GetBalanceExpression),
	}

	return &Application{
		expressions: top,
		repo:        repo,
	}
}

type BotMention struct {
	Platform string
	BotID    string
	UserID   string
	Text     string
}

type BotResponse struct {
	Text string
}

func (a Application) ProcessAppMention(ctx context.Context, m *BotMention) BotResponse {
	//botId := viper.GetString("BOT_USER_ID")

	// Extract captures of the first match. If no matches then return generic "idk lol" message.
	for i := range a.expressions {
		if a.expressions[i].MatchString(m.Text) {
			captures := extractNamedCaptures(a.expressions[i], m.Text)
			// Do other processing.
			log.Printf("Captures: %+#v\n", captures)
			s := a.processCommand(ctx, m, captures)
			return BotResponse{Text: s}
		}
	}

	// Generic "idk lol" message
	return BotResponse{Text: GenericResponse}
}

func (a Application) processCommand(ctx context.Context, m *BotMention, captures map[string]string) string {
	grantSubCommand := "grant[[:space:]]*(?P<currency>[$A-Za-z]+).*"
	r := regexp.MustCompile(grantSubCommand)
	if r.MatchString(captures["command"]) {
		commandCaptures := extractNamedCaptures(r, captures["command"])
		log.Printf("Command Captures: %+v\n", commandCaptures)

		granter, err := a.repo.GetOrCreateUserBySlackID(ctx, cleanSlackUserID(m.UserID))
		if err != nil {
			return GenericErrorResponse
		}
		recipient, err := a.repo.GetOrCreateUserBySlackID(ctx, cleanSlackUserID(captures["recipient_id"]))
		if err != nil {
			return GenericErrorResponse
		}

		_, err = a.repo.GrantCurrency(ctx, commandCaptures["currency"], granter.ID, recipient.ID)
		if err != nil {
			return GenericErrorResponse
		}
		return fmt.Sprintf("Huzzah! Successfully granted 1.00 %s to %s", commandCaptures["currency"], captures["recipient_id"])
	}

	return GenericResponse
}

func cleanSlackUserID(id string) string {
	replacer := strings.NewReplacer("<", "", ">", "", "@", "")
	return replacer.Replace(id)
}

func extractNamedCaptures(e *regexp.Regexp, input string) map[string]string {
	match := e.FindStringSubmatch(input)
	captures := make(map[string]string, len(match))

	for i, name := range e.SubexpNames() {
		if i > 0 && i <= len(match) {
			captures[name] = match[i]
		}
	}

	return captures
}
