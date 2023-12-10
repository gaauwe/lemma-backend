package database

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

type Inbox struct {
	Enabled     bool       `json:"enabled" binding:"required"`
	LastChecked *time.Time `json:"lastChecked"`
}

type WatcherFilter struct {
	Title   string `json:"title"`
	Author  string `json:"author"`
	Upvotes int64  `json:"upvotes"`
	Link    string `json:"link"`
}

type Watcher struct {
	ID          string        `json:"id"`
	Name        string        `json:"name" binding:"required"`
	Community   string        `json:"community" binding:"required"`
	Filters     WatcherFilter `json:"filters"`
	LastChecked *time.Time    `json:"lastChecked"`
}

type User struct {
	Username    string             `json:"username" binding:"required"`
	Token       string             `json:"token" binding:"required"`
	DeviceToken string             `json:"deviceToken" binding:"required"`
	Inbox       *Inbox             `json:"inbox,omitempty"`
	Watchers    map[string]Watcher `json:"watchers"`
}

func GetUserByUsername(username string) (User, error) {
	db := Get()

	// Fetch user from the DB.
	doc, err := db.FindFirst(query.NewQuery("users").Where(query.Field("Username").Eq(username)))
	if err != nil || doc == nil {
		return User{}, errors.New("User could not be found")
	}

	// Map document to user struct.
	user := User{}
	doc.Unmarshal(user)

	return user, nil
}

func GetUsers() ([]User, error) {
	db := Get()

	// Fetch all users from the DB.
	docs, err := db.FindAll(query.NewQuery("users"))
	if err != nil {
		return []User{}, errors.New("Users could not be retrieved")
	}

	// Map all the documents to a user struct.
	users := []User{}
	for _, doc := range docs {
		user := User{}
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

	// Check if watcher exists.
	exists, err := db.Exists(query.NewQuery("users").Where(query.Field(key).Exists()))
	log.Println(exists)
	if err != nil {
		return Watcher{}, err
	}
	if exists == false {
		return Watcher{}, errors.New("Watcher not found")
	}

	// Create update map to store in the DB.
	data := make(map[string]interface{})
	data[key] = watcher

	err = db.Update(query.NewQuery("users").Where(query.Field("Username").Eq(username)), data)
	return watcher, err
}

func DeleteWatcher(username string, id string) error {
	key := fmt.Sprintf("Watchers.%s", id)
	db := Get()

	// Check if watcher exists.
	exists, err := db.Exists(query.NewQuery("users").Where(query.Field(key).Exists()))
	log.Println(exists)
	if err != nil {
		return err
	}
	if exists == false {
		return errors.New("Watcher not found")
	}

	err = db.UpdateFunc(query.NewQuery("users").Where(query.Field("Username").Eq(username)), func(doc *document.Document) *document.Document {
		watchers := doc.Get("Watchers")

		// Make sure that Watchers is a map.
		if m, ok := watchers.(map[string]interface{}); ok {
			// Remove the watcher from watchers and update the document with the remaining watchers.
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
