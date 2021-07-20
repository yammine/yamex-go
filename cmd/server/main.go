package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/yammine/yamex-go/adapter"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	if err := repo.Migrate(); err != nil {
		log.Fatalf("failed to migrate db: %s", err)
	}
	user, err := repo.GetOrCreateUserBySlackID(context.Background(), "new_slack_id2")
	if err != nil {
		log.Fatalf("failed GetOrCreateUserBySlackID %s", err)
	}
	fmt.Printf("User: %+v\n", user)

	r := gin.Default()
	r.POST("/events", func(c *gin.Context) {
		var req ChallengeReq

		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Println("the error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("req", req)
		c.JSON(http.StatusOK, gin.H{"challenge": req.Challenge})
	})

	r.Run(":3000")
}
