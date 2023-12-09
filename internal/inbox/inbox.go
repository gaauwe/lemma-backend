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

func FetchReplies(c *lemmy.Client, ctx context.Context, username string) {
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

			if !shouldSkipEvent(reply.Comment.Published, username) {
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

	// Update the last checked of this user, so we never send notifications again for events that happened before this moment.
	database.UpdateUserInboxLastChecked(username)
}

func shouldSkipEvent(time types.LemmyTime, username string) bool {
	user, err := database.GetUserByUsername(username)
	if err != nil {
		return true
	}

	// If we never checked this user before, we rely on the time that the server is started.
	if user.Inbox.LastChecked == nil {
		return config.Get().Server.StartedAt.After(time.Time)
	}

	return user.Inbox.LastChecked.After(time.Time)
}
