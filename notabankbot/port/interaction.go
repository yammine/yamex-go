package port

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

type SlackInteractor struct {
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

func NewSlackInteractor() *SlackInteractor {
	return &SlackInteractor{}
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

		if err := r.ParseForm(); err != nil {
			fmt.Println("error parsing form")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse payload into our struct
		type wrapper struct {
			Payload []*SlackInteraction `json:"payload"`
		}
		res := &wrapper{}
		payload := []byte(r.Form.Get("payload"))
		if err := json.Unmarshal(payload, &res); err != nil {
			w.WriteHeader(500)
			fmt.Println("error unmarshalling", err)
			fmt.Println("raw", string(payload))
			fmt.Println("form", r.Form)
			return
		}

		// Business logic
		for i := range res.Payload {
			interaction := res.Payload[i]
			fmt.Println("interaction", interaction)
		}

		w.WriteHeader(200)
	}
}

func (s SlackInteractor) ProcessInteraction(i *SlackInteraction) error {
	return nil
}
