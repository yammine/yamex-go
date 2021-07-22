package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/yammine/yamex-go/notabankbot/adapter"
	"github.com/yammine/yamex-go/notabankbot/app"
	"github.com/yammine/yamex-go/notabankbot/port"
)

type ChallengeReq struct {
	Challenge string `json:"challenge"`
}

func main() {
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Crashing due to failed config read: %s", err)
	}

	// Playing with postgres adapter
	repo := adapter.NewPostgresRepository(viper.GetString("POSTGRES_DSN"))
	fmt.Println(repo)
	application := &app.Application{}
	slackConsumer := port.NewSlackConsumer(application)

	router := mux.NewRouter()
	router.HandleFunc("/slack/events", slackConsumer.Handler())
	router.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		log.Println(string(body))
		writer.WriteHeader(200)
		return
	})

	srv := &http.Server{
		Addr:         "0.0.0.0:3000",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
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
	log.Println("shutting down")
	os.Exit(0)
}
