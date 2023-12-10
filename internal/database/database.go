package database

import (
	"log"

	"github.com/ostafen/clover/v2"
)

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
