package main

import (
	"log"
	"time"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/inbox"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"github.com/go-co-op/gocron"
)

func main() {
	err := config.LoadConfig("./settings.toml")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	err = notification.NewClient()
	if err != nil {
		log.Fatal("Failed to read auth key from file: ", err)
	}

	s := gocron.NewScheduler(time.UTC)
	_, _ = s.Every(5).Seconds().Do(func() { inbox.FetchReplies() })
	s.StartBlocking()
}
