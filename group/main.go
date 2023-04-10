package main

import (
	"context"
	"fmt"
	"github.com/gabrielseibel1/gaef/auth"
	"github.com/gabrielseibel1/gaef/client/user"
	"github.com/gabrielseibel1/gaef/group/handler"
	"github.com/gabrielseibel1/gaef/group/store"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	// read environment variables
	port := os.Getenv("PORT")
	userServiceURL := os.Getenv("USER_SERVICE_URL")
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
	a := auth.NewMiddlewareGenerator(user.Client{URL: userServiceURL}, "userID", "token")
	s := store.New(client.Database(dbName).Collection(collectionName))
	h := handler.New(s, s, s, s, s, s, s)
	handlers := handlers{
		auth:                    a,
		onlyLeaders:             h,
		createGroup:             h,
		readParticipatingGroups: h,
		readLeadingGroups:       h,
		readGroup:               h,
		updateGroup:             h,
		deleteGroup:             h,
	}

	// run http server
	server := gin.Default()
	groups := server.Group("/api/v0/groups")
	groups.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	authed := groups.Group("", handlers.auth.AuthMiddleware())
	{
		authed.POST("/", handlers.createGroup.CreateGroupHandler())
		authed.GET("/participating", handlers.readParticipatingGroups.ReadParticipatingGroupsHandler())
		authed.GET("/leading", handlers.readLeadingGroups.ReadLeadingGroupsHandler())
		authed.GET("/:id", handlers.readGroup.ReadGroupHandler())

		forLeaders := authed.Group("", handlers.onlyLeaders.OnlyLeadersMiddleware())
		{
			forLeaders.GET("/leading/:id", handlers.readGroup.ReadGroupHandler())
			forLeaders.PUT("/:id", handlers.updateGroup.UpdateGroupHandler())
			forLeaders.DELETE("/:id", handlers.deleteGroup.DeleteGroupHandler())
		}
	}
	log.Fatal(server.Run(fmt.Sprintf("0.0.0.0:%s", port)))
}

type handlers struct {
	auth                    AuthMiddleware
	onlyLeaders             OnlyLeadersMiddleware
	createGroup             CreateGroupHandler
	readParticipatingGroups ReadParticipatingGroupsHandler
	readLeadingGroups       ReadLeadingGroupsHandler
	readGroup               ReadGroupHandler
	updateGroup             UpdateGroupHandler
	deleteGroup             DeleteGroupHandler
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
type ReadParticipatingGroupsHandler interface {
	ReadParticipatingGroupsHandler() gin.HandlerFunc
}
type ReadLeadingGroupsHandler interface {
	ReadLeadingGroupsHandler() gin.HandlerFunc
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
