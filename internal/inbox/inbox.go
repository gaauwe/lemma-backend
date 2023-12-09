package inbox

import (
	"context"
	"fmt"
	"log"

	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"go.elara.ws/go-lemmy"
	"go.elara.ws/go-lemmy/types"
)

func FetchReplies(c *lemmy.Client, ctx context.Context) {
	user, err := c.UnreadCount(ctx, types.GetUnreadCount{
		Auth: c.Token,
	})
	if err != nil {
		log.Fatal("Error:", err)
	}

	var title string
	var body string
	var image string
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
				title = fmt.Sprintf("New reply from %s", author)
				body = reply.Comment.Content
				image = reply.Creator.Avatar.String()
			}
		}
	}

	if len(title) > 0 && len(body) > 0 {
		notification.SendNotification(title, body, image, count)
	}
}
