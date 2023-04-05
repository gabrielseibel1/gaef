package main

import (
	"context"
	"fmt"
	"github.com/gabrielseibel1/gaef/auth"
	"github.com/gabrielseibel1/gaef/client/group"
	"github.com/gabrielseibel1/gaef/client/user"
	"github.com/gabrielseibel1/gaef/encounter-proposal/api"
	"github.com/gabrielseibel1/gaef/encounter-proposal/store"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type handlerGenerators struct {
	authMiddlewareGenerator                        authMiddlewareGenerator
	epCreatorGroupLeaderCheckerMiddlewareGenerator epCreatorGroupLeaderMiddlewareGenerator
	epCreationHandlerGenerator                     epCreationHandlerGenerator
	epReadingAllHandlerGenerator                   epReadingAllHandlerGenerator
	epReadingByUserHandlerGenerator                epReadingByUserHandlerGenerator
	epReadingByIDHandlerGenerator                  epReadingByIDHandlerGenerator
	epUpdateHandlerGenerator                       epUpdateHandlerGenerator
	epDeletionHandlerGenerator                     epDeletionHandlerGenerator
	appCreationHandlerGenerator                    appCreationHandlerGenerator
}

type authMiddlewareGenerator interface {
	AuthMiddleware() gin.HandlerFunc
}
type epCreatorGroupLeaderMiddlewareGenerator interface {
	EPCreatorGroupLeaderCheckerMiddleware() gin.HandlerFunc
}
type epCreationHandlerGenerator interface {
	EPCreationHandler() gin.HandlerFunc
}
type epReadingAllHandlerGenerator interface {
	EPReadingAllHandler() gin.HandlerFunc
}
type epReadingByUserHandlerGenerator interface {
	EPReadingByUserHandler() gin.HandlerFunc
}
type epReadingByIDHandlerGenerator interface {
	EPReadingByIDHandler() gin.HandlerFunc
}
type epUpdateHandlerGenerator interface {
	EPUpdateHandler() gin.HandlerFunc
}
type epDeletionHandlerGenerator interface {
	EPDeletionHandler() gin.HandlerFunc
}
type appCreationHandlerGenerator interface {
	AppCreationHandler() gin.HandlerFunc
}

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
	authHandler := auth.NewMiddlewareGenerator(userClient, "userID", "token")
	db := store.New(client.Database(dbName).Collection(collectionName))
	encounterProposalsAPI := api.New(db, db, db, db, db, db, db, groupClient, groupClient)
	hg := handlerGenerators{
		authMiddlewareGenerator:                        authHandler,
		epCreatorGroupLeaderCheckerMiddlewareGenerator: encounterProposalsAPI,
		epCreationHandlerGenerator:                     encounterProposalsAPI,
		epReadingAllHandlerGenerator:                   encounterProposalsAPI,
		epReadingByUserHandlerGenerator:                encounterProposalsAPI,
		epReadingByIDHandlerGenerator:                  encounterProposalsAPI,
		epUpdateHandlerGenerator:                       encounterProposalsAPI,
		epDeletionHandlerGenerator:                     encounterProposalsAPI,
		appCreationHandlerGenerator:                    encounterProposalsAPI,
	}

	// run HTTP server
	server := gin.Default()
	root := server.Group("/api/v0/encounter-proposals", hg.authMiddlewareGenerator.AuthMiddleware())
	{
		root.POST("/", hg.epCreationHandlerGenerator.EPCreationHandler())
		root.GET("/page/:"+api.Page, hg.epReadingAllHandlerGenerator.EPReadingAllHandler())
		root.GET("/mine", hg.epReadingByUserHandlerGenerator.EPReadingByUserHandler())

		byEPID := root.Group("/:" + api.EPID)
		{
			byEPID.GET("", hg.epReadingByIDHandlerGenerator.EPReadingByIDHandler())

			creatorsOnly := byEPID.Group("", hg.epCreatorGroupLeaderCheckerMiddlewareGenerator.EPCreatorGroupLeaderCheckerMiddleware())
			{
				creatorsOnly.PUT("", hg.epUpdateHandlerGenerator.EPUpdateHandler())
				creatorsOnly.DELETE("", hg.epDeletionHandlerGenerator.EPDeletionHandler())
			}

			byEPID.POST("/applications", hg.appCreationHandlerGenerator.AppCreationHandler())
		}
	}
	log.Fatal(server.Run(fmt.Sprintf("0.0.0.0:%s", port)))
}
