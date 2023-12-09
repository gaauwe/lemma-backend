package user

import (
	"context"
	"log"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/inbox"
	"github.com/gaauwe/lemma-backend/internal/watcher"
	"go.elara.ws/go-lemmy"
)

func CheckNotifications() {
	users, err := database.GetUsers()
	if err != nil {
		log.Fatal("Failed to get users: ", err)
	}

	for _, user := range users {
		ctx := context.Background()
		c, err := lemmy.New(config.Get().Lemmy.Server)
		c.Token = user.Token

		if err != nil {
			log.Println("Failed to initialize Lemmy client: ", err)
			return
		}

		inbox.FetchReplies(c, ctx, user.Username)
		watcher.FetchPosts(c, ctx)
	}
}
