package inbox

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/lemmy"
	"github.com/gaauwe/lemma-backend/internal/lemmy/types"
	"github.com/gaauwe/lemma-backend/internal/notification"
)

func FetchReplies(c *lemmy.Client, ctx context.Context, user *database.User) {
	unread, err := c.UnreadCount(ctx, types.GetUnreadCount{
		Auth: c.Token,
	})
	if err != nil {
		log.Println("Failed to retrieve unread count: ", err)

		// If the token is for some reason not valid, we delete the user and notify them of the issue.
		// TODO: Update user in the database, so that we only send this once.
		if strings.Contains(err.Error(), "not_logged_in") {
			notification.SendNotification("Something went wrong with fetching your notifications", "Please disable and re-enable your notifications in the Lemma settings", "", 1, "/TODO", user)
		}
		return
	}

	var title string
	var body string
	var image string
	var url string
	var commentReplyId int64
	count := unread.Replies

	log.Println("Unread notifications:", count)
	if count > 0 {
		replies, err := c.Replies(ctx, types.GetReplies{
			Auth:       c.Token,
			UnreadOnly: types.NewOptional(true),
		})
		if err != nil {
			log.Println("Failed to retrieve replies: ", err)
		}

		if len(replies.Replies) > 0 {
			reply := replies.Replies[0]

			if !shouldSkipEvent(reply.Comment.Published, user.Username) {
				author := reply.Creator.Name
				title = fmt.Sprintf("New reply from %s", author)
				body = reply.Comment.Content
				image = reply.Creator.Avatar.String()
				url = fmt.Sprintf("/post/%d?commentPath=%d", reply.Post.ID, reply.Comment.ID)
				commentReplyId = reply.CommentReply.ID
			}
		}
	}

	if len(title) > 0 && len(body) > 0 {
		data := make(map[string]interface{})
		data["commentReplyId"] = commentReplyId

		log.Println(title)
		notification.SendNotification(title, body, image, count, url, user)
	}

	// Update the last checked of this user, so we never send notifications again for events that happened before this moment.
	database.UpdateUserInboxLastChecked(user.Username)
}

func shouldSkipEvent(time types.LemmyTime, username string) bool {
	user, err := database.GetUserByUsername(username)
	if err != nil {
		return true
	}

	/**
	 * We rely on the time that the server was started in the following cases:
	 * - We never checked this user before
	 * - The server was restarted since the last time we checked this user
	 */
	if user.Inbox.LastChecked == nil || user.Inbox.LastChecked.Before(config.Get().Server.StartedAt) {
		return config.Get().Server.StartedAt.After(time.Time)
	}

	return user.Inbox.LastChecked.After(time.Time)
}
