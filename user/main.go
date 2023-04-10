package main

import (
	"context"
	"fmt"
	"github.com/gabrielseibel1/gaef/user/hasher"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gabrielseibel1/gaef/user/handler"
	"github.com/gabrielseibel1/gaef/user/store"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// dependencies

type AuthHandler interface {
	JWTAuthMiddleware() gin.HandlerFunc
}
type SignupHandler interface {
	Signup() gin.HandlerFunc
}
type LoginHandler interface {
	Login() gin.HandlerFunc
}
type TokenHandler interface {
	GetIDFromToken() gin.HandlerFunc
}
type GetHandler interface {
	GetUserFromID() gin.HandlerFunc
}
type UpdateHandler interface {
	UpdateUser() gin.HandlerFunc
}
type DeleteHandler interface {
	DeleteUser() gin.HandlerFunc
}

type handlerGenerator struct {
	authHandler   AuthHandler
	signupHandler SignupHandler
	loginHandler  LoginHandler
	tokenHandler  TokenHandler
	getHandler    GetHandler
	updateHandler UpdateHandler
	deleteHandler DeleteHandler
}

// implementation

func main() {
	// read environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	port := os.Getenv("PORT")
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
	str := store.NewMongoStore(client.Database(dbName).Collection(collectionName))
	phv := hasher.New()
	hdl := handler.New(phv, phv, str, str, str, str, str, []byte(jwtSecret))
	gen := handlerGenerator{
		authHandler:   hdl,
		signupHandler: hdl,
		loginHandler:  hdl,
		tokenHandler:  hdl,
		getHandler:    hdl,
		updateHandler: hdl,
		deleteHandler: hdl,
	}

	r := gin.Default()
	users := r.Group("/api/v0/users")
	{
		users.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

		public := users.Group("")
		{
			public.POST("/", gen.signupHandler.Signup())
			public.POST("/session", gen.loginHandler.Login())
		}
		auth := users.Group("", gen.authHandler.JWTAuthMiddleware())
		{
			auth.GET("/token-validation", gen.tokenHandler.GetIDFromToken())
			auth.GET("/:id", gen.getHandler.GetUserFromID())
			auth.PUT("/:id", gen.updateHandler.UpdateUser())
			auth.DELETE("/:id", gen.deleteHandler.DeleteUser())
		}
	}
	log.Fatal(r.Run(fmt.Sprintf("0.0.0.0:%s", port)))
}
