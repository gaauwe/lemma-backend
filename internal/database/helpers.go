package database

import (
	"errors"
	"time"

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
	Watchers    []Watcher
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

func AddWatcher(username string, watcher Watcher) error {
	user, err := GetUserByUsername(username)
	if err != nil {
		return err
	}

	// Loop through all the current watchers to prepare the DB update.
	watchers := []map[string]interface{}{}
	for _, w := range user.Watchers {
		// Prevent multiple watchers with the same name.
		if watcher.Name == w.Name {
			return errors.New("Watcher with that name already exists")
		}

		watchers = append(watchers, watcherToMap(w))
	}

	// Add new watcher to the list.
	watchers = append(watchers, watcherToMap(watcher))

	err = updateWatchers(username, watchers)
	return err
}

func EditWatcher(username string, watcher Watcher) error {
	user, err := GetUserByUsername(username)
	if err != nil {
		return err
	}

	// Loop through all the current watchers to prepare the DB update.
	watchers := []map[string]interface{}{}
	for _, w := range user.Watchers {
		// Replace the watcher that we want to edit.
		if watcher.ID == w.ID {
			watchers = append(watchers, watcherToMap(watcher))
		} else {
			watchers = append(watchers, watcherToMap(w))
		}
	}

	err = updateWatchers(username, watchers)
	return err
}

func DeleteWatcher(username string, id string) error {
	user, err := GetUserByUsername(username)
	if err != nil {
		return err
	}

	// Loop through all the current watchers to prepare the DB update.
	watchers := []map[string]interface{}{}
	for _, w := range user.Watchers {
		// Filter out the watcher that we want to delete.
		if id != w.ID {
			watchers = append(watchers, watcherToMap(w))
		}
	}

	err = updateWatchers(username, watchers)
	return err
}

func UpdateWatcherLastChecked(username string, watcher Watcher) error {
	t := time.Now()
	watcher.LastChecked = &t
	return EditWatcher(username, watcher)
}

func updateWatchers(username string, watchers []map[string]interface{}) error {
	db := Get()
	data := make(map[string]interface{})
	data["Watchers"] = watchers

	err := db.Update(query.NewQuery("users").Where(query.Field("Username").Eq(username)), data)
	return err
}

func watcherToMap(watcher Watcher) map[string]interface{} {
	result := make(map[string]interface{})

	result["ID"] = watcher.ID
	result["Name"] = watcher.Name
	result["Community"] = watcher.Community
	result["Filters"] = watcherFilterToMap(watcher.Filters)
	result["LastChecked"] = watcher.LastChecked

	return result
}

func watcherFilterToMap(filter WatcherFilter) map[string]interface{} {
	result := make(map[string]interface{})

	result["Title"] = filter.Title
	result["Author"] = filter.Author
	result["Upvotes"] = filter.Upvotes
	result["Link"] = filter.Link

	return result
}
