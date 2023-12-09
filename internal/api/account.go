package api

import (
	"net/http"

	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gaauwe/lemma-backend/internal/notification"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	var newUser database.User
	if err := ctx.BindJSON(&newUser); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}

	// Check if the device token is valid.
	err := notification.SendRegistrationNotification(newUser.DeviceToken)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid device token"})
		return
	}

	// Check if the username is already registered.
	db := database.Get()
	existingUser, _ := db.Exists(query.NewQuery("users").Where(query.Field("Username").Eq(newUser.Username)))
	if existingUser {
		ctx.IndentedJSON(http.StatusUnprocessableEntity, gin.H{"message": "User already exists"})
		return
	}

	// Create document for new user.
	doc := document.NewDocumentOf(newUser)

	// Add document to the users collection.
	_, err = db.InsertOne("users", doc)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "User could not be added"})
		return
	}

	ctx.IndentedJSON(http.StatusCreated, newUser)
}

func DeleteUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	db := database.Get()

	// Remove user from the DB.
	err := db.Delete(query.NewQuery("users").Where(query.Field("Username").Eq(username)))
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "User could not be deleted"})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "User succesfully deleted"})
}

func AddWatcher(ctx *gin.Context) {
	username := ctx.Param("username")
	var newWatcher database.Watcher
	if err := ctx.BindJSON(&newWatcher); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}

	// Generate UUID for the new watcher, which can be used to delete the watcher later on.
	id := uuid.New()
	newWatcher.ID = id.String()

	// Add watcher to the user in the DB.
	err := database.AddWatcher(username, newWatcher)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, newWatcher)
}

func EditWatcher(ctx *gin.Context) {
	username := ctx.Param("username")
	var newWatcher database.Watcher
	if err := ctx.BindJSON(&newWatcher); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid body"})
		return
	}

	err := database.EditWatcher(username, newWatcher)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Watcher succesfully updated"})
}

func DeleteWatcher(ctx *gin.Context) {
	username := ctx.Param("username")
	id := ctx.Param("id")

	err := database.DeleteWatcher(username, id)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{"message": "Watcher succesfully deleted"})
}
