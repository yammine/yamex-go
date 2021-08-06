package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/yammine/yamex-go/notabankbot/domain"
)

const (
	// Top Level expressions

	CommandExpression    = "(?P<bot_id><@[A-Z0-9]{11}>)(?P<command>.+)(?P<recipient_id><@[A-Z0-9]{11}>)(?P<note>.*)"
	GetBalanceExpression = "(?P<bot_id><@[A-Z0-9]{11}>)[[:space:]](balance|my[[:space:]]balance)"

	// Sub-command expressions

	GrantCurrencyExpression = "grant[[:space:]]*(?P<currency>[$A-Za-z]+).*"

	// Command names

	GetBalanceCmd = "GetBalance"
	CommandCmd    = "Command"

	// Sub-command Names

	GrantCurrencyCmd = "GrantCurrency"

	// Responses

	GenericResponse      = "I don't understand what you're asking me :face_with_head_bandage:"
	GenericErrorResponse = "I seem to be experiencing an unexpected error :robot_face:"

	AlreadyGrantedCurrency = "Oops! Looks like you've already granted currency recently. Try again later :simple_smile:"

	// Capture Keys

	ckRecipientID = "recipient_id"
	ckCurrency    = "currency"
	ckNote        = "note"
	ckCommand     = "command"
)

type Application struct {
	expressions           map[string]*regexp.Regexp
	subCommandExpressions map[string]*regexp.Regexp
	repo                  Repository
}

func NewApplication(repo Repository) *Application {
	top := map[string]*regexp.Regexp{
		CommandCmd:    regexp.MustCompile(CommandExpression),
		GetBalanceCmd: regexp.MustCompile(GetBalanceExpression),
	}

	sub := map[string]*regexp.Regexp{
		GrantCurrencyCmd: regexp.MustCompile(GrantCurrencyExpression),
	}

	return &Application{
		expressions:           top,
		subCommandExpressions: sub,
		repo:                  repo,
	}
}

type BotMention struct {
	Platform string
	UserID   string
	Text     string
}

type BotResponse struct {
	Text string
}

func (a Application) ProcessAppMention(ctx context.Context, m *BotMention) BotResponse {
	//botId := viper.GetString("BOT_USER_ID")

	for name, expression := range a.expressions {
		if expression.MatchString(m.Text) {
			captures := extractNamedCaptures(expression, m.Text)
			// Do other processing.
			log.Printf("Captures: %+#v\n", captures)
			var response string
			switch name {
			case CommandCmd:
				response = a.processCommand(ctx, m, captures)
			case GetBalanceCmd:
				response = a.GetBalance(ctx, m)
			}

			return BotResponse{Text: response}
		}
	}

	// Generic "idk lol" message
	return BotResponse{Text: GenericResponse}
}

func (a Application) Grant(ctx context.Context, m *BotMention, captures map[string]string) string {
	granter, err := a.repo.GetOrCreateUserBySlackID(ctx, cleanSlackUserID(m.UserID))
	if err != nil {
		return GenericErrorResponse
	}
	recipient, err := a.repo.GetOrCreateUserBySlackID(ctx, cleanSlackUserID(captures[ckRecipientID]))
	if err != nil {
		return GenericErrorResponse
	}

	input := GrantCurrencyInput{
		Currency:   captures[ckCurrency],
		FromUserID: granter.ID,
		ToUserID:   recipient.ID,
		Note:       strings.TrimSpace(captures[ckNote]),
	}
	_, err = a.repo.GrantCurrency(ctx, input)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyGranted) {
			return AlreadyGrantedCurrency
		}
		return GenericErrorResponse
	}
	return fmt.Sprintf("Huzzah! Successfully granted 1.00 %s to <@%s> with reason: %s", captures[ckCurrency], recipient.SlackID, captures[ckNote])
}

func (a Application) GetBalance(ctx context.Context, m *BotMention) string {
	user, err := a.repo.GetOrCreateUserBySlackID(ctx, m.UserID)
	if err != nil {
		return GenericErrorResponse
	}
	accounts, err := a.repo.GetAccountsForUser(ctx, user.ID)
	tableData := make([][]string, len(accounts))
	for i := range accounts {
		acc := accounts[i]
		tableData[i] = []string{fmt.Sprint(acc.ID), acc.Currency, acc.Balance.StringFixed(8)}
	}

	buf := bytes.NewBuffer([]byte{})
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Account ID", "Currency", "Balance"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, v := range tableData {
		table.Append(v)
	}
	table.Render()

	return fmt.Sprintf("Here are all of your accounts on record:\n```%s```", buf.String())
}

func (a Application) processCommand(ctx context.Context, m *BotMention, captures map[string]string) string {
	for name, expression := range a.subCommandExpressions {
		if expression.MatchString(captures[ckCommand]) {
			// Get captures within command, merge them into existing map
			cmdCaptures := extractNamedCaptures(expression, captures[ckCommand])
			for k, v := range cmdCaptures {
				captures[k] = v
			}

			// Find out which func to call
			switch name {
			case GrantCurrencyCmd:
				return a.Grant(ctx, m, captures)
			default:
				return GenericResponse
			}
		}
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
