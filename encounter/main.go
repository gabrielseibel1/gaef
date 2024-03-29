package main

import (
	"context"
	"fmt"
	"github.com/gabrielseibel1/gaef/auth"
	"github.com/gabrielseibel1/gaef/client/group"
	"github.com/gabrielseibel1/gaef/client/user"
	"github.com/gabrielseibel1/gaef/encounter/api"
	"github.com/gabrielseibel1/gaef/encounter/server"
	"github.com/gabrielseibel1/gaef/encounter/store"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// read environment variables
	port := os.Getenv("PORT")
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	groupServiceURL := os.Getenv("GROUP_SERVICE_URL")
	dbURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DATABASE")
	collectionName := os.Getenv("MONGODB_COLLECTION")

	// connect to mongoDB
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(dbURI).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
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
	userClient := user.Client{URL: userServiceURL}
	groupClient := group.Client{URL: groupServiceURL}
	mongoStore := store.New(client.Database(dbName).Collection(collectionName))
	authentication := auth.NewMiddlewareGenerator(userClient, "userID", "token")
	apis := api.New(groupClient, mongoStore, mongoStore, mongoStore, mongoStore, mongoStore, mongoStore, mongoStore)
	handlers := server.New(apis, apis, apis, apis, apis, apis, apis)

	// setup HTTP server
	app := gin.Default()
	encounters := app.Group("/api/v0/encounters")
	encounters.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	authed := encounters.Group("", authentication.AuthMiddleware())
	{
		noID := authed.Group("/")
		{
			noID.GET("", handlers.ReadUserEncountersHandler())
			noID.POST("", handlers.CreateEncounterHandler())
		}
		byID := authed.Group("/:" + server.EncIDParam)
		{
			byID.GET("", handlers.ReadEncounterHandler())
			byID.PUT("", handlers.UpdateEncounterHandler())
			byID.DELETE("", handlers.DeleteEncounterHandler())

			confirmation := byID.Group("/confirmation")
			{
				confirmation.POST("", handlers.ConfirmEncounterHandler())
				confirmation.DELETE("", handlers.DeclineEncounterHandler())
			}
		}
	}
	log.Fatal(app.Run(fmt.Sprintf("0.0.0.0:%s", port)))
}
