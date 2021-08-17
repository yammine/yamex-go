package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/slack-go/slack"

	"github.com/rs/zerolog"

	"github.com/gorilla/mux"
	_ "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/yammine/yamex-go/notabankbot/adapter"
	"github.com/yammine/yamex-go/notabankbot/app"
	"github.com/yammine/yamex-go/notabankbot/port"
)

const ServiceName = "yamex"

func main() {
	if !viper.IsSet("PRODUCTION") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	viper.AutomaticEnv()
	viper.SetDefault("PORT", 3000)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("viper couldn't find config.yml, falling back to ENV config")
	}

	// Playing with postgres adapter
	repo := adapter.NewPostgresRepository(viper.GetString("POSTGRES_DSN"))
	repo.Migrate()
	application := app.NewApplication(repo)
	slackConsumer := port.NewSlackConsumer(application)

	router := mux.NewRouter()
	router.HandleFunc("/slack/events", slackConsumer.Handler())
	router.HandleFunc("/slack/oauth", oAuthRedirectHandler(repo))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("PORT")),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Str("service", ServiceName).Msg("error starting http listener")
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	log.Info().Msg("Shutting down")
	os.Exit(0)
}

func oAuthRedirectHandler(repo app.Repository) func(w http.ResponseWriter, r *http.Request) {
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
