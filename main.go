package main

import (
	"log"
	"time"

	"github.com/gaauwe/lemma-backend/internal/api"
	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"github.com/gaauwe/lemma-backend/internal/user"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func main() {
	// Load config.
	err := config.LoadConfig("./settings.toml")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Setup APN notification client.
	err = notification.SetupClient()
	if err != nil {
		log.Fatal("Failed to read auth key from file: ", err)
	}

	// Setup the database.
	err = database.SetupClient()
	if err != nil {
		log.Fatal("Failed to setup database: ", err)
	}

	// Start checking notifications periodically.
	pollRate := config.Get().Server.PollRate
	if pollRate > 0 {
		s := gocron.NewScheduler(time.UTC)
		_, err = s.Every(int(pollRate)).Seconds().Do(func() { user.CheckNotifications() })
		if err != nil {
			log.Fatal("Failed to setup cron job: ", err)
		}

		s.StartAsync()
	}

	// Register API routes.
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	protected := router.Group("/api")
	protected.Use(api.AuthMiddleware())

	protected.POST("/users", api.PostUsers)

	protected.POST("/users/:username/watcher", api.AddWatcher)
	protected.PUT("/users/:username/watcher/:id", api.EditWatcher)
	protected.DELETE("/users/:username/watcher/:id", api.DeleteWatcher)

	protected.PUT("/users/:username/inbox", api.EditInbox)

	// TODO: Remove these API routes in production, or add authorization.
	protected.GET("/users", api.GetUsers)
	protected.GET("/users/:username", api.GetUserByUsername)
	protected.DELETE("/users/:username", api.DeleteUserByUsername)

	if config.Get().Server.EnableSSL {
		log.Fatal(autotls.Run(router, "gromdroid.nl"))
	} else {
		router.Run()
	}
}
