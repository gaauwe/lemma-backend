package user

import (
	"context"
	"fmt"
	"log"
	"strings"

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
		server := fmt.Sprintf("https://%s", strings.Split(user.Username, "@")[1])

		ctx := context.Background()
		c, err := lemmy.New(server)
		c.Token = user.Token

		if err != nil {
			log.Println("Failed to initialize Lemmy client: ", err)
			return
		}

		inbox.FetchReplies(c, ctx, user)
		watcher.FetchPosts(c, ctx, user)
	}
}
