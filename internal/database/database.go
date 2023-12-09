package database

import (
	"log"
	"strconv"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/ostafen/clover/v2"
	"go.elara.ws/go-lemmy/types"
)

var dbOld *badger.DB
var db *clover.DB

func SetupClient() error {
	database, err := clover.Open("database")
	if err != nil {
		return err
	}

	db = database
	collectionExists, err := db.HasCollection("users")
	if err != nil {
		return err
	}

	if !collectionExists {
		log.Println("Creating new users collection...")
		err := db.CreateCollection("users")
		if err != nil {
			return err
		}
	}

	return nil
}

func Get() *clover.DB {
	return db
}

func SetupClientOld() error {
	database, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		return err
	}

	dbOld = database

	// Store last checked when initializing the database, so we can skip all previous notifications.
	StoreLastChecked()
	return nil
}

func StoreLastChecked() {
	err := dbOld.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("last_checked"), []byte(strconv.FormatInt((time.Now().Unix()), 10)))
		return err
	})
	if err != nil {
		log.Fatal("Failed to store last checked in DB: ", err)
	}
}

func IsAfterLastChecked(date types.LemmyTime) bool {
	var lastChecked []byte
	err := dbOld.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("last_checked"))
		if err != nil {
			log.Fatal("Failed to get last_checked from DB: ", err)
		}

		lastChecked, err = item.ValueCopy(nil)
		if err != nil {
			log.Fatal("Failed to get last_checked from DB: ", err)
		}

		return nil
	})
	if err != nil {
		return false
	}

	unixTimestamp, err := strconv.ParseInt(string(lastChecked), 10, 64)
	if err != nil {
		log.Fatal("Failed to convert last_checked to timestamp: ", err)
	}

	timestamp := time.Unix(unixTimestamp, 0)
	StoreLastChecked()
	return timestamp.Before(date.Time)
}
