package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yammine/yamex-go/adapter"

	f "github.com/fauna/faunadb-go/v4/faunadb"
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

	client := f.NewFaunaClient(viper.GetString("FAUNA_SECRET"), f.Endpoint("https://db.us.fauna.com"), f.HTTP(&http.Client{}))
	if err := adapter.InitFaunaDatabase(client); err != nil {
		log.Fatalf("Error initializing FaunaDB: %s", err)
	}

	repo := adapter.NewFaunaRepository(client)
	user, err := repo.GetOrCreateUser("testingasdfasdfasd")
	fmt.Printf("Fetched user: %+v\n", user)
	if err != nil {
		log.Fatalf("error creating user: %s", err)
	}

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
