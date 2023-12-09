package watcher

import (
	"context"
	"log"

	"github.com/gaauwe/lemma-backend/internal/notification"
	"go.elara.ws/go-lemmy"
	"go.elara.ws/go-lemmy/types"
)

func FetchPosts(c *lemmy.Client, ctx context.Context) {
	posts, err := c.Posts(ctx, types.GetPosts{
		Auth:          types.NewOptional(c.Token),
		CommunityName: types.NewOptional("lemma@lemmy.world"),
		Sort:          types.NewOptional(types.SortTypeNew),
	})
	if err != nil {
		log.Println("Failed to retrieve posts: ", err)
	}

	var title string
	var body string
	var image string
	var count int64

	if len(posts.Posts) > 0 {
		post := posts.Posts[0]
		title = post.Post.Name
		body = post.Post.Body.String()
		image = post.Post.ThumbnailURL.String()
	}

	if len(title) > 0 && len(body) > 0 {
		notification.SendNotification(title, body, image, count)
	}
}
