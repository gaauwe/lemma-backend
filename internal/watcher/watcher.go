package watcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"go.elara.ws/go-lemmy"
	"go.elara.ws/go-lemmy/types"
)

func FetchPosts(c *lemmy.Client, ctx context.Context, username string) {
	user, err := database.GetUserByUsername(username)
	if err != nil {
		log.Println("Failed to retrieve watchers for user", username, ": ", err)
		return
	}

	for _, watcher := range user.Watchers {
		lastChecked := watcher.LastChecked
		posts, err := c.Posts(ctx, types.GetPosts{
			Auth:          types.NewOptional(c.Token),
			CommunityName: types.NewOptional(watcher.Community),
			Sort:          types.NewOptional(types.SortTypeNew),
		})
		if err != nil {
			log.Println("Failed to retrieve posts for community", watcher.Community, ":", err)
		}

		var title string
		var body string
		var image string
		var count int64

		if len(posts.Posts) > 0 && !shouldSkipEvent(posts.Posts[0].Post.Published, lastChecked) {
			post := posts.Posts[0]
			title = fmt.Sprintf("New post in %s", post.Community.Name)
			body = post.Post.Name
			image = post.Post.ThumbnailURL.String()
		}

		if len(title) > 0 && len(body) > 0 {
			notification.SendNotification(title, body, image, count)
		}

		// Update the last checked of this user, so we never send notifications again for events that happened before this moment.
		database.UpdateWatcherLastChecked(username, watcher)
	}
}

func shouldSkipEvent(time types.LemmyTime, lastChecked *time.Time) bool {
	// If we never checked this user before, we rely on the time that the server is started.
	if lastChecked == nil {
		return config.Get().Server.StartedAt.After(time.Time)
	}

	return lastChecked.After(time.Time)
}
