package port

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/viper"

	"github.com/yammine/yamex-go/notabankbot/app"
)

const (
	// Top Level expressions

	CommandExpression    = "(?P<bot_id><@[A-Z0-9]{11}>)(?P<command>.+)(?P<recipient_id><@[A-Z0-9]{11}>)(?P<note>.*)"
	GetBalanceExpression = "(?P<bot_id><@[A-Z0-9]{11}>)[[:space:]](balance|my[[:space:]]balance)"
	FeedbackExpression   = "(?P<bot_id><@[A-Z0-9]{11}>)[[:space:]]+feedback.*"

	// Sub-command expressions

	GrantCurrencyExpression = "grant[[:space:]]*(?P<currency>[$A-Za-z]+).*"
	GetBalanceForExpression = "(get balance|balance for).*"
	SendCurrencyExpression  = "send[[:space:]]+(?P<amount>[-+]?[0-9]*\\.?[0-9]*)[[:space:]]+(?P<currency>[$A-Za-z]+)"

	// Command names

	GetBalanceCmd = "GetBalance"
	CommandCmd    = "Command"
	FeedbackCmd   = "Feedback"

	// Sub-command Names

	GrantCurrencyCmd = "GrantCurrency"
	GetBalanceForCmd = "GetBalanceFor"
	SendCurrencyCmd  = "SendCurrency"
)

type SlackConsumer struct {
	app         *app.Application
	credentials SlackCredentialStore

	expressions           map[string]*regexp.Regexp
	subCommandExpressions map[string]*regexp.Regexp
}

func NewSlackConsumer(app *app.Application, credentialRepo SlackCredentialStore) *SlackConsumer {
	top := map[string]*regexp.Regexp{
		CommandCmd:    regexp.MustCompile(CommandExpression),
		GetBalanceCmd: regexp.MustCompile(GetBalanceExpression),
		FeedbackCmd:   regexp.MustCompile(FeedbackExpression),
	}

	sub := map[string]*regexp.Regexp{
		GrantCurrencyCmd: regexp.MustCompile(GrantCurrencyExpression),
		GetBalanceForCmd: regexp.MustCompile(GetBalanceForExpression),
		SendCurrencyCmd:  regexp.MustCompile(SendCurrencyExpression),
	}

	return &SlackConsumer{
		app:                   app,
		credentials:           credentialRepo,
		expressions:           top,
		subCommandExpressions: sub,
	}
}

func (s SlackConsumer) Handler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv, err := slack.NewSecretsVerifier(r.Header, viper.GetString("SLACK_SIGNING_SECRET"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var resp *slackevents.ChallengeResponse
			err := json.Unmarshal(body, &resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(resp.Challenge))
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent

			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				log.Debug().Dict(
					"AppMentionEvent",
					zerolog.Dict().
						Str("type", ev.Type).
						Str("user", ev.User).
						Str("text", ev.Text),
				).Msg("event received")

				response := s.ProcessAppMention(ctx, &BotMention{
					Platform: "slack",
					UserID:   ev.User,
					Text:     replaceWhitespace(ev.Text),
				})
				// TODO: Move this to somewhere else
				token, err := s.credentials.GetCredentials(ctx, eventsAPIEvent.TeamID)
				if err != nil {
					log.Error().Err(err).Msg("Failed to get slack credentials")
					w.WriteHeader(500)
					return
				}
				client := slack.New(token)
				go s.reply(client, ev, response)

			default:

			}
		}
	}
}

func (s SlackConsumer) reply(client *slack.Client, ev *slackevents.AppMentionEvent, response BotResponse) {
	_, _, _, err := client.SendMessage(
		ev.Channel,
		slack.MsgOptionText(response.Text, false),
		slack.MsgOptionTS(messageTS(ev)),
		slack.MsgOptionBlocks(response.Blocks...),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to send response")
	}
}

func messageTS(ev *slackevents.AppMentionEvent) string {
	// If ev.ThreadTimeStamp is set then use that (reply in thread mention was made).
	// Else use TimeStamp (start new threaded reply)
	if ev.ThreadTimeStamp != "" {
		return ev.ThreadTimeStamp
	} else {
		return ev.TimeStamp
	}
}

func replaceWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
}
