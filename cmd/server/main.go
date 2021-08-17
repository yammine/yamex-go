package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

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
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	viper.AutomaticEnv()
	//viper.SetConfigName("config")
	//viper.SetConfigType("yml")
	//viper.AddConfigPath(".")
	//if err := viper.ReadInConfig(); err != nil {
	//	log.Fatal().Err(err).Str("service", ServiceName).Msgf("cannot start %s", ServiceName)
	//}

	// Playing with postgres adapter
	repo := adapter.NewPostgresRepository(viper.GetString("POSTGRES_DSN"))
	repo.Migrate()
	application := app.NewApplication(repo)
	slackConsumer := port.NewSlackConsumer(application)

	router := mux.NewRouter()
	router.HandleFunc("/slack/events", slackConsumer.Handler())

	srv := &http.Server{
		Addr:         "0.0.0.0:3000",
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
