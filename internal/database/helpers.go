package database

import (
	"errors"
	"time"

	"github.com/ostafen/clover/v2/query"
)

type Inbox struct {
	LastChecked *time.Time
}

type User struct {
	ID          string
	Username    string
	Token       string
	DeviceToken string
	Inbox       Inbox
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
	user.ID = doc.ObjectId()

	return user, nil
}

func UpdateUserInboxLastChecked(username string) error {
	db := Get()

	data := make(map[string]interface{})
	data["Inbox.LastChecked"] = time.Now()

	err := db.Update(query.NewQuery("users").Where(query.Field("Username").Eq(username)), data)
	return err
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
		user.ID = doc.ObjectId()
		users = append(users, user)
	}

	return users, nil
}
