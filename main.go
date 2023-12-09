package main

import (
	"log"
	"time"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"github.com/gaauwe/lemma-backend/internal/user"
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
	err = database.SetupClient()
	if err != nil {
		log.Fatal("Failed to setup database: ", err)
	}

	// Start checking notifications periodically.
	s := gocron.NewScheduler(time.UTC)
	_, _ = s.Every(config.Get().Server.PollRate).Seconds().Do(func() { user.CheckNotifications() })
	s.StartBlocking()
}
