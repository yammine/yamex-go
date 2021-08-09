package port

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/rs/zerolog"

	"github.com/yammine/yamex-go/notabankbot/domain"

	"github.com/olekukonko/tablewriter"
	_ "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/yammine/yamex-go/notabankbot/app"
)

type BotMention struct {
	Platform string
	UserID   string
	Text     string
}

func (b BotMention) MarshalZerologObject(e *zerolog.Event) {
	e.Str("UserID", b.UserID).Str("Platform", b.Platform).Str("Text", b.Text)
}

var _ zerolog.LogObjectMarshaler = (*BotMention)(nil)

type BotResponse struct {
	Text string
}

func (s SlackConsumer) ProcessAppMention(ctx context.Context, m *BotMention) BotResponse {
	//botId := viper.GetString("BOT_USER_ID")

	for name, expression := range s.expressions {
		if expression.MatchString(m.Text) {
			captures := extractNamedCaptures(expression, m.Text)
			// Do other processing.
			var response string
			switch name {
			case CommandCmd:
				response = s.processCommand(ctx, m, captures)
			case GetBalanceCmd:
				response = s.processGetBalanceQuery(ctx, m.UserID)
			}

			return BotResponse{Text: response}
		}
	}

	log.Error().Str("text", m.Text).Msg("Could not process mention")
	return BotResponse{Text: GenericResponse}
}

func (s SlackConsumer) processGetBalanceQuery(ctx context.Context, slackUserID string) string {
	accounts, err := s.app.GetBalance(ctx, &app.GetBalanceInput{UserID: slackUserID})
	if err != nil {
		log.Error().Err(err).Msg("Error processing GetBalance query")
		return GenericErrorResponse
	}

	accountsTable := renderAccounts(accounts)

	return fmt.Sprintf("Here are your account balances:\n```%s```", accountsTable)
}

func (s SlackConsumer) processCommand(ctx context.Context, m *BotMention, captures map[string]string) string {
	command := captures[ckCommand]

	for name, expression := range s.subCommandExpressions {
		if expression.MatchString(command) {
			// Get captures within command, merge them into existing map
			cmdCaptures := extractNamedCaptures(expression, command)
			for k, v := range cmdCaptures {
				captures[k] = v
			}

			// Find out which func to call
			switch name {
			case GetBalanceForCmd:
				return s.processGetBalanceQuery(ctx, cleanSlackUserID(captures[ckRecipientID]))
			case GrantCurrencyCmd:
				_, err := s.app.Grant(ctx, &app.GrantInput{
					GranterID:  cleanSlackUserID(m.UserID),
					ReceiverID: cleanSlackUserID(captures[ckRecipientID]),
					Platform:   "slack",
					Currency:   captures[ckCurrency],
					Note:       captures[ckNote],
				})

				if err != nil {
					log.Error().Object("context", m).Err(err).Msg("Error granting currency")
					if errors.Is(err, domain.ErrAlreadyGranted) {
						return AlreadyGrantedCurrency
					}
					return GenericErrorResponse
				}

				return fmt.Sprintf("Success! Granted 1 %s to %s. Spend it wisely :sunglasses:", captures[ckCurrency], captures[ckRecipientID])
			case SendCurrencyCmd:
				amount, err := decimal.NewFromString(captures[ckAmount])
				if err != nil {
					log.Error().Err(err).Object("context", m).Msg("could not parse amount from message")
					return GenericErrorResponse
				}

				err = s.app.Transfer(ctx, &app.TransferInput{
					SenderID:   cleanSlackUserID(m.UserID),
					ReceiverID: cleanSlackUserID(captures[ckRecipientID]),
					Platform:   "slack",
					Currency:   captures[ckCurrency],
					Note:       captures[ckNote],
					Amount:     amount,
				})

				if err != nil {
					log.Error().Err(err).Object("context", m).Msg("Error during transfer")
					return GenericErrorResponse
				}
				return fmt.Sprintf("Success! Sent %s %s to %s for reason: `%s`.\n\nThanks for using yamex!", amount.String(), captures[ckCurrency], captures[ckRecipientID], strings.TrimSpace(captures[ckNote]))
			default:
				log.Error().Object("context", m).Str("name", name).Str("command", command).Msg("Could not match command")
				return GenericResponse
			}
		}
	}

	log.Error().Str("command", command).Object("context", m).Msg("Could not match command")
	return GenericResponse
}

func renderAccounts(accounts []*domain.Account) string {
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

	return buf.String()
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
