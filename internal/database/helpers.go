package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

type Inbox struct {
	Enabled     bool
	LastChecked *time.Time
}

type WatcherFilter struct {
	Title   string
	Author  string
	Upvotes int64
	Link    string
}

type Watcher struct {
	ID          string
	Name        string
	Community   string
	Filters     WatcherFilter
	LastChecked *time.Time
}

type User struct {
	Username    string
	Token       string
	DeviceToken string
	Inbox       Inbox
	Watchers    map[string]Watcher
}

func GetUserByUsername(username string) (*User, error) {
	db := Get()

	// Fetch user from the DB.
	doc, err := db.FindFirst(query.NewQuery("users").Where(query.Field("Username").Eq(username)))
	if err != nil || doc == nil {
		return &User{}, errors.New("User could not be found")
	}

	// Map document to user struct.
	user := &User{}
	doc.Unmarshal(user)

	return user, nil
}

func GetUsers() ([]*User, error) {
	db := Get()

	// Fetch all users from the DB.
	docs, err := db.FindAll(query.NewQuery("users"))
	if err != nil {
		return []*User{}, errors.New("Users could not be retrieved")
	}

	// Map all the documents to a user struct.
	users := []*User{}
	for _, doc := range docs {
		user := &User{}
		doc.Unmarshal(user)
		users = append(users, user)
	}

	return users, nil
}

func UpdateUserInboxLastChecked(username string) error {
	db := Get()

	data := make(map[string]interface{})
	data["Inbox.LastChecked"] = time.Now()

	err := db.Update(query.NewQuery("users").Where(query.Field("Username").Eq(username)), data)
	return err
}

func UpdateUserInboxEnabled(username string, enabled bool) error {
	_, err := GetUserByUsername(username)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["Inbox.Enabled"] = enabled

	err = db.Update(query.NewQuery("users").Where(query.Field("Username").Eq(username)), data)
	return err
}

func AddWatcher(username string, watcher Watcher) (Watcher, error) {
	// Generate UUID for the new watcher.
	id := uuid.New()
	key := fmt.Sprintf("Watchers.%s", id)
	watcher.ID = id.String()

	db := Get()
	data := make(map[string]interface{})
	data[key] = watcher

	err := db.Update(query.NewQuery("users").Where(query.Field("Username").Eq(username)), data)
	return watcher, err
}

func EditWatcher(username string, id string, watcher Watcher) (Watcher, error) {
	key := fmt.Sprintf("Watchers.%s", id)

	db := Get()
	data := make(map[string]interface{})
	data[key] = watcher

	err := db.Update(query.NewQuery("users").Where(query.Field("Username").Eq(username)), data)
	return watcher, err
}

func DeleteWatcher(username string, id string) error {
	db := Get()
	err := db.UpdateFunc(query.NewQuery("users").Where(query.Field("Username").Eq(username)), func(doc *document.Document) *document.Document {
		watchers := doc.Get("Watchers")

		// Make sure that Watchers is a map.
		if m, ok := watchers.(map[string]interface{}); ok {
			// Remove the watcher from watchers and update the document with the new watchers.
			delete(m, id)
			doc.Set("Watchers", m)
		}

		return doc
	})
	return err
}

func UpdateWatcherLastChecked(username string, watcher Watcher) (Watcher, error) {
	t := time.Now()
	watcher.LastChecked = &t
	return EditWatcher(username, watcher.ID, watcher)
}
