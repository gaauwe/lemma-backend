package main

import (
	"log"
	"time"

	"github.com/gaauwe/lemma-backend/internal/api"
	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"github.com/gaauwe/lemma-backend/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func main() {
	// Load config.
	err := config.LoadConfig("./settings.toml")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Setup APN notification client.
	err = notification.SetupClient()
	if err != nil {
		log.Fatal("Failed to read auth key from file: ", err)
	}

	// Setup the database.
	err = database.SetupClientOld()
	if err != nil {
		log.Fatal("Failed to setup old database: ", err)
	}

	err = database.SetupClient()
	if err != nil {
		log.Fatal("Failed to setup database: ", err)
	}

	// Start checking notifications periodically.
	pollRate := config.Get().Server.PollRate
	if pollRate > 0 {
		s := gocron.NewScheduler(time.UTC)
		_, _ = s.Every(pollRate).Seconds().Do(func() { user.CheckNotifications() })
		s.StartBlocking()
	}

	// Register API routes.
	router := gin.Default()
	router.GET("/users", api.GetUsers)
	router.GET("/users/:username", api.GetUserByUsername)
	router.POST("/users", api.PostUsers)
	router.DELETE("/users/:username", api.DeleteUserByUsername)

	router.Run(config.Get().Server.Addr)
}
