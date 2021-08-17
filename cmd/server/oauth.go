package main

import (
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/yammine/yamex-go/notabankbot/app"

	"github.com/spf13/viper"

	"github.com/slack-go/slack"
)

func OAuthRedirectHandler(repo app.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		client := &http.Client{Timeout: 5 * time.Second}
		oauthResp, err := slack.GetOAuthV2Response(client, viper.GetString("SLACK_CLIENT_ID"), viper.GetString("SLACK_CLIENT_SECRET"), code, viper.GetString("SLACK_REDIRECT_URI"))
		spew.Dump(oauthResp)
		if err != nil {
			// TODO: something
		}

		// Persist the token & team ID, so we can use them later when responding to mentions
		if err := repo.SaveCredentials(r.Context(), oauthResp.Team.ID, oauthResp.AccessToken); err != nil {
			// TODO: something
		}

		w.WriteHeader(200)
		return
	}
}
