package port

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/yammine/yamex-go/notabankbot/app"

	"github.com/rs/zerolog/log"

	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

type SlackInteractor struct {
	app         *app.Application
	credentials SlackCredentialStore
}

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Container struct {
	Type        string `json:"type"`
	MessageTS   string `json:"message_ts"`
	ChannelID   string `json:"channel_id"`
	IsEphemeral bool   `json:"is_ephemeral"`
	ThreadTS    string `json:"thread_ts"`
}

type Action struct {
	Type     string `json:"type"`
	ActionID string `json:"action_id"`
	BlockID  string `json:"block_id"`
	Value    string `json:"value"`
	ActionTS string `json:"action_ts"`
}

type SlackInteraction struct {
	Type    string      `json:"type"`
	User    *slack.User `json:"user"`
	Channel *Channel    `json:"channel"`
	Team    *slack.Team `json:"team"`

	Token       string `json:"token"`
	ResponseURL string `json:"response_url"`
	TriggerID   string `json:"trigger_id"`

	Actions []*Action `json:"actions"`
}

func NewSlackInteractor(credentials SlackCredentialStore, app *app.Application) *SlackInteractor {
	return &SlackInteractor{
		app:         app,
		credentials: credentials,
	}
}

func (s SlackInteractor) Handler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the body & check contents are actually from Slack
		body, err := io.ReadAll(r.Body)
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
			fmt.Println("error verifying r body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Parse form data so we can eventaully reach that json payload
		form, err := url.ParseQuery(string(body))
		if err != nil {
			fmt.Println("couldn't parse form body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse payload into our struct
		res := &SlackInteraction{}
		payload := []byte(form.Get("payload"))
		if err := json.Unmarshal(payload, &res); err != nil {
			w.WriteHeader(500)
			fmt.Println("error unmarshalling", err)
			return
		}

		// Business logic
		fmt.Printf("SlackInteraction: %+v\n", res)
		if err := s.ProcessInteraction(res); err != nil {
			fmt.Println("failed to process actions", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
	}
}

func (s SlackInteractor) ProcessInteraction(i *SlackInteraction) error {
	// setup
	token, err := s.credentials.GetCredentials(context.Background(), i.Team.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get slack credentials")
		return err
	}
	client := slack.New(token)
	var response string

	// Process value

	// Reply
	s.respondToAction(client, i, response)

	return nil
}

func (s SlackInteractor) respondToAction(client *slack.Client, i *SlackInteraction, response string) {
	_, _, _, err := client.SendMessage(
		i.Channel.ID,
		slack.MsgOptionText(response, false),
		slack.MsgOptionResponseURL(i.ResponseURL, "ephemeral"),
		slack.MsgOptionReplaceOriginal(i.ResponseURL),
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to respond to action")
	}
}
