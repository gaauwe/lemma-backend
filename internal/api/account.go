package api

import (
	"log"
	"net/http"

	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"github.com/gin-gonic/gin"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

func GetUsers(ctx *gin.Context) {
	users, err := database.GetUsers()
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, users)
}

func GetUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	user, err := database.GetUserByUsername(username)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, user)
}

func PostUsers(ctx *gin.Context) {
	var user database.User
	if err := ctx.BindJSON(&user); err != nil {
		handleValidationErrors(ctx, err)
		return
	}

	// Check if the device token is valid, by sendig a silent notification.
	err := notification.SendRegistrationNotification(user.DeviceToken)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid device token"})
		return
	}

	// Check if the username is already registered.
	db := database.Get()
	existingUser, _ := db.Exists(query.NewQuery("users").Where(query.Field("Username").Eq(user.Username)))
	if existingUser {
		log.Println(err)

		ctx.IndentedJSON(http.StatusUnprocessableEntity, gin.H{"message": "User already exists"})
		return
	}

	// Add missing data if necessary.
	if user.Inbox == nil {
		user.Inbox = &database.Inbox{
			Enabled: false,
		}
	}

	// Create document for new user.
	doc := document.NewDocumentOf(user)

	// Add document to the users collection.
	_, err = db.InsertOne("users", doc)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusCreated, user)
}

func DeleteUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	db := database.Get()

	// Remove user from the DB.
	err := db.Delete(query.NewQuery("users").Where(query.Field("Username").Eq(username)))
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "User succesfully deleted"})
}

func EditInbox(ctx *gin.Context) {
	username := ctx.Param("username")

	var inbox database.Inbox
	if err := ctx.BindJSON(&inbox); err != nil {
		handleValidationErrors(ctx, err)
		return
	}

	// Update notification settings in the DB.
	err := database.UpdateUserInboxEnabled(username, inbox.Enabled)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Inbox notitification settings updated succesfully"})
}

func AddWatcher(ctx *gin.Context) {
	username := ctx.Param("username")

	var watcher database.Watcher
	if err := ctx.BindJSON(&watcher); err != nil {
		handleValidationErrors(ctx, err)
		return
	}

	// Add watcher to the user in the DB.
	result, err := database.AddWatcher(username, watcher)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, result)
}

func EditWatcher(ctx *gin.Context) {
	username := ctx.Param("username")
	id := ctx.Param("id")

	var watcher database.Watcher
	if err := ctx.BindJSON(&watcher); err != nil {
		handleValidationErrors(ctx, err)
		return
	}

	// Edit watcher in the DB.
	result, err := database.EditWatcher(username, id, watcher)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, result)
}

func DeleteWatcher(ctx *gin.Context) {
	username := ctx.Param("username")
	id := ctx.Param("id")

	// Delete watcher in the DB.
	err := database.DeleteWatcher(username, id)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Watcher succesfully deleted"})
}
