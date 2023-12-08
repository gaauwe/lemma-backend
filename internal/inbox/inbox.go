package inbox

import (
	"context"
	"fmt"
	"log"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"go.elara.ws/go-lemmy"
	"go.elara.ws/go-lemmy/types"
)

func FetchReplies() {
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

	user, err := c.UnreadCount(ctx, types.GetUnreadCount{
		Auth: c.Token,
	})
	if err != nil {
		log.Fatal("Error:", err)
	}

	var title string
	var body string
	count := user.Replies

	if count > 0 {
		replies, err := c.Replies(ctx, types.GetReplies{
			Auth:       c.Token,
			UnreadOnly: types.NewOptional(true),
		})
		if err != nil {
			log.Fatal("Error:", err)
		}

		if len(replies.Replies) > 0 {
			reply := replies.Replies[0]

			if database.IsAfterLastChecked(reply.Comment.Published) {
				author := reply.Creator.Name
				post := reply.Post.Name
				title = fmt.Sprintf("%s replied to your comment in %s", author, post)
				body = reply.Comment.Content
			}
		}
	}

	if len(title) > 0 && len(body) > 0 {
		notification.SendNotification(title, body, count)
	} else {
		log.Printf("No new notifications...")
	}
}
