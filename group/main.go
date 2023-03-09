package main

import (
	"context"
	"fmt"
	"gaef-group-service/auth"
	"gaef-group-service/handler"
	"gaef-group-service/store"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	// read environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	dbURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DATABASE")
	collectionName := os.Getenv("MONGODB_COLLECTION")

	// TODO: secure connection to mongo with user/password
	// connect to mongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	// instantiate and inject dependencies
	auth := auth.New(userServiceURL)
	store := store.New(client.Database(dbName).Collection(collectionName))
	handler := handler.New(auth, store, store, store, store, store, store)
	handlers := handlers{
		auth:          handler,
		onlyLeaders:   handler,
		createGroup:   handler,
		readAllGroups: handler,
		readGroup:     handler,
		updateGroup:   handler,
		deleteGroup:   handler,
	}

	// run http server
	server := gin.Default()
	api := server.Group("/api/v0/groups", handlers.auth.AuthMiddleware())
	{
		api.POST("/", handlers.createGroup.CreateGroupHandler())
		api.GET("/", handlers.readAllGroups.ReadAllGroupsHandler())
		api.GET("/:id", handlers.readGroup.ReadGroupHandler())

		forLeaders := api.Group("", handlers.onlyLeaders.OnlyLeadersMiddleware())
		{
			forLeaders.PUT("/:id", handlers.updateGroup.UpdateGroupHandler())
			forLeaders.DELETE("/:id", handlers.deleteGroup.DeleteGroupHandler())
		}
	}
	server.Run(fmt.Sprintf("0.0.0.0:%s", port))
}

type handlers struct {
	auth          AuthMiddleware
	onlyLeaders   OnlyLeadersMiddleware
	createGroup   CreateGroupHandler
	readAllGroups ReadAllGroupsHandler
	readGroup     ReadGroupHandler
	updateGroup   UpdateGroupHandler
	deleteGroup   DeleteGroupHandler
}

type AuthMiddleware interface {
	AuthMiddleware() gin.HandlerFunc
}
type OnlyLeadersMiddleware interface {
	OnlyLeadersMiddleware() gin.HandlerFunc
}
type CreateGroupHandler interface {
	CreateGroupHandler() gin.HandlerFunc
}
type ReadAllGroupsHandler interface {
	ReadAllGroupsHandler() gin.HandlerFunc
}
type ReadGroupHandler interface {
	ReadGroupHandler() gin.HandlerFunc
}
type UpdateGroupHandler interface {
	UpdateGroupHandler() gin.HandlerFunc
}
type DeleteGroupHandler interface {
	DeleteGroupHandler() gin.HandlerFunc
}
