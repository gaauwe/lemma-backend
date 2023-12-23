package main

import (
	"log"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/gaauwe/lemma-backend/internal/api"
	"github.com/gaauwe/lemma-backend/internal/config"
	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"github.com/gaauwe/lemma-backend/internal/user"
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

	// User routes.
	protectedUser := router.Group("/api/users/:username")
	protectedUser.Use(api.UserAuthMiddleware())

	protectedUser.POST("/watcher", api.AddWatcher)
	protectedUser.PUT("/watcher/:id", api.EditWatcher)
	protectedUser.DELETE("/watcher/:id", api.DeleteWatcher)

	protectedUser.PUT("/inbox", api.EditInbox)
	protectedUser.PUT("/token", api.EditDeviceToken)

	// Admin routes.
	protectedAdmin := router.Group("/api")
	protectedAdmin.Use(api.AdminAuthMiddleware())

	protectedAdmin.POST("/users", api.PostUsers)

	// TODO: Remove these API routes in production, or add authorization.
	protectedAdmin.GET("/users", api.GetUsers)
	protectedAdmin.GET("/users/:username", api.GetUserByUsername)
	protectedAdmin.DELETE("/users/:username", api.DeleteUserByUsername)

	if config.Get().Server.EnableSSL {
		certmagic.DefaultACME.Agreed = true
		certmagic.DefaultACME.Email = config.Get().Server.Email
		err = certmagic.HTTPS([]string{config.Get().Server.Domain}, router)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		router.Run()
	}
}
