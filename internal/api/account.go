package api

import (
	"log"
	"net/http"

	"github.com/gaauwe/lemma-backend/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

type UserRequest struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type User struct {
	UserRequest
	ID string `json:"id"`
}

var users = []UserRequest{
	{Username: "gromdroid", Token: "1"},
}

func GetUsers(ctx *gin.Context) {
	db := database.Get()

	// Fetch all users from the DB.
	docs, err := db.FindAll(query.NewQuery("users"))
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "User could not be retrieved"})
		return
	}

	// Map all the documents to a user struct.
	users := []*User{}
	for _, doc := range docs {
		user := &User{}
		doc.Unmarshal(user)
		user.ID = doc.ObjectId()
		users = append(users, user)
	}

	ctx.IndentedJSON(http.StatusOK, users)
}

func GetUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	db := database.Get()

	// Fetch user from the DB.
	doc, err := db.FindFirst(query.NewQuery("users").Where(query.Field("Username").Eq(username)))
	if err != nil {
		log.Println(err)
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "User could not be retrieved"})
		return
	}

	// Return 404 if no user is found.
	if doc == nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Map document to user struct.
	user := &User{}
	doc.Unmarshal(user)
	user.ID = doc.ObjectId()

	ctx.IndentedJSON(http.StatusOK, user)
}

func PostUsers(ctx *gin.Context) {
	var newUser User
	if err := ctx.BindJSON(&newUser); err != nil {
		return
	}

	// Check if the username is already registered
	db := database.Get()
	existingUser, _ := db.Exists(query.NewQuery("users").Where(query.Field("Username").Eq(newUser.Username)))
	if existingUser {
		ctx.IndentedJSON(http.StatusUnprocessableEntity, gin.H{"message": "User already exists"})
		return
	}

	// Create document for new user.
	doc := document.NewDocumentOf(newUser)

	// Add document to the users collection.
	id, err := db.InsertOne("users", doc)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "User could not be added"})
		return
	}

	newUser.ID = id
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
