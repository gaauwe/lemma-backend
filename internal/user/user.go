package user

import (
	"context"
	"log"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/inbox"
	"github.com/gaauwe/lemma-backend/internal/watcher"
	"go.elara.ws/go-lemmy"
	"go.elara.ws/go-lemmy/types"
)

func CheckNotifications() {
	ctx := context.Background()

	c, err := lemmy.New(config.Get().Lemmy.Server)
	if err != nil {
		log.Fatal("Error:", err)
	}

	err = c.ClientLogin(ctx, types.Login{
		UsernameOrEmail: config.Get().Lemmy.Username,
		Password:        config.Get().Lemmy.Password,
	})
	if err != nil {
		log.Fatal("Error:", err)
	}

	inbox.FetchReplies(c, ctx, "gromdroid")
	watcher.FetchPosts(c, ctx)
}
