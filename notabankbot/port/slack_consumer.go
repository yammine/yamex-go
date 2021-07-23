package port

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/viper"

	"github.com/yammine/yamex-go/notabankbot/app"
)

type SlackConsumer struct {
	app    *app.Application
	client *slack.Client
}

func NewSlackConsumer(app *app.Application) *SlackConsumer {
	return &SlackConsumer{
		app:    app,
		client: slack.New(viper.GetString("BOT_USER_OAUTH_TOKEN")),
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
				log.Printf("AppMentionEvent: %+v", ev)
				response := s.app.ProcessAppMention(ctx, &app.BotMention{
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

func (s SlackConsumer) reply(ev *slackevents.AppMentionEvent, response app.BotResponse) {
	s.client.SendMessage(
		ev.Channel,
		slack.MsgOptionText(response.Text, false),
		slack.MsgOptionTS(ev.TimeStamp),
	)
}
