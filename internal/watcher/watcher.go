package watcher

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/lemmy"
	"github.com/gaauwe/lemma-backend/internal/lemmy/types"
	"github.com/gaauwe/lemma-backend/internal/notification"
)

func FetchPosts(c *lemmy.Client, ctx context.Context, user *database.User) {
	for _, watcher := range user.Watchers {
		log.Println("Checking community:", watcher.Community)

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
		var url string

		if len(posts.Posts) > 0 && !shouldSkipEvent(posts.Posts[0], watcher) {
			post := posts.Posts[0]
			title = fmt.Sprintf("New post in %s", post.Community.Name)
			body = post.Post.Name
			image = post.Post.ThumbnailURL.String()
			url = fmt.Sprintf("/post/%d", post.Post.ID)
		}

		if len(title) > 0 && len(body) > 0 {
			notification.SendNotification(title, body, image, count, url, user)
		}

		// Update the last checked of this user, so we never send notifications again for events that happened before this moment.
		database.UpdateWatcherLastChecked(user.Username, watcher)
	}
}

func shouldSkipEvent(post types.PostView, watcher database.Watcher) bool {
	if happenedInThePast(post.Post.Published, watcher.LastChecked) {
		return true
	}

	if titleMatches(post, watcher) &&
		authorMatches(post, watcher) &&
		upvotesMatches(post, watcher) &&
		linkMatches(post, watcher) {
		return false
	}

	return true
}

func happenedInThePast(time types.LemmyTime, lastChecked *time.Time) bool {
	/**
	 * We rely on the time that the server was started in the following cases:
	 * - We never checked this watcher before
	 * - The server was restarted since the last time we checked this watcher
	 */
	if lastChecked == nil || lastChecked.Before(config.Get().Server.StartedAt) {
		return config.Get().Server.StartedAt.After(time.Time)
	}

	return lastChecked.After(time.Time)
}

func titleMatches(post types.PostView, watcher database.Watcher) bool {
	if len(watcher.Filters.Title) == 0 {
		return true
	}

	title := strings.ToLower(post.Post.Name)
	filter := strings.ToLower(watcher.Filters.Title)

	return strings.Contains(title, filter)
}

func authorMatches(post types.PostView, watcher database.Watcher) bool {
	if len(watcher.Filters.Author) == 0 {
		return true
	}

	author := strings.ToLower(getUserFullName(post.Creator))
	filter := strings.ToLower(watcher.Filters.Author)

	return author == filter
}

func upvotesMatches(post types.PostView, watcher database.Watcher) bool {
	if watcher.Filters.Upvotes == 0 {
		return true
	}

	return post.Counts.Upvotes > watcher.Filters.Upvotes
}

func linkMatches(post types.PostView, watcher database.Watcher) bool {
	if len(watcher.Filters.Link) == 0 {
		return true
	}

	link := strings.ToLower(post.Post.URL.String())
	filter := strings.ToLower(watcher.Filters.Link)

	return strings.Contains(link, filter)
}

func getUserFullName(person types.Person) string {
	return fmt.Sprintf("%s@%s", person.Name, getBaseUrl(person.ActorID))
}

func getBaseUrl(link string) string {
	regex := regexp.MustCompile(`^(?:https?:\/\/)?([^/]+)`)
	matches := regex.FindStringSubmatch(link)

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
