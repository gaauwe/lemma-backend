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
		_, err = s.Every(int(pollRate)).Seconds().Do(func() { user.CheckNotifications() })
		if err != nil {
			log.Fatal("Failed to setup cron job: ", err)
		}

		s.StartAsync()
	}

	// Register API routes.
	router := gin.Default()
	router.POST("/users", api.PostUsers)

	router.POST("/users/:username/watcher", api.AddWatcher)
	router.PUT("/users/:username/watcher/:id", api.EditWatcher)
	router.DELETE("/users/:username/watcher/:id", api.DeleteWatcher)

	// TODO: Remove these API routes in production, or add authorization.
	router.GET("/users", api.GetUsers)
	router.GET("/users/:username", api.GetUserByUsername)
	router.DELETE("/users/:username", api.DeleteUserByUsername)

	router.Run(config.Get().Server.Addr)
}
