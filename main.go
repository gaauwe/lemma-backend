package main

import (
	"log"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/inbox"
	"github.com/gaauwe/lemma-backend/internal/notification"
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

	inbox.FetchReplies()
}
