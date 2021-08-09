package port

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

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

	// Sub-command expressions

	GrantCurrencyExpression = "grant[[:space:]]*(?P<currency>[$A-Za-z]+).*"
	GetBalanceForExpression = "(get balance|balance for).*"

	// Command names

	GetBalanceCmd = "GetBalance"
	CommandCmd    = "Command"

	// Sub-command Names

	GrantCurrencyCmd = "GrantCurrency"
	GetBalanceForCmd = "GetBalanceFor"

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

type SlackConsumer struct {
	app    *app.Application
	client *slack.Client

	expressions           map[string]*regexp.Regexp
	subCommandExpressions map[string]*regexp.Regexp
}

func NewSlackConsumer(app *app.Application) *SlackConsumer {
	top := map[string]*regexp.Regexp{
		CommandCmd:    regexp.MustCompile(CommandExpression),
		GetBalanceCmd: regexp.MustCompile(GetBalanceExpression),
	}

	sub := map[string]*regexp.Regexp{
		GrantCurrencyCmd: regexp.MustCompile(GrantCurrencyExpression),
		GetBalanceForCmd: regexp.MustCompile(GetBalanceForExpression),
	}

	return &SlackConsumer{
		app:                   app,
		client:                slack.New(viper.GetString("BOT_USER_OAUTH_TOKEN")),
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
					Text:     ev.Text,
				})

				go s.reply(ev, response)

			default:

			}
		}
	}
}

func (s SlackConsumer) reply(ev *slackevents.AppMentionEvent, response BotResponse) {
	s.client.SendMessage(
		ev.Channel,
		slack.MsgOptionText(response.Text, false),
		slack.MsgOptionTS(ev.TimeStamp),
	)
}
