package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gabrielseibel1/gaef-user-service/domain"
	"github.com/gabrielseibel1/gaef-user-service/handler"
	"github.com/gabrielseibel1/gaef-user-service/service"
	"github.com/gabrielseibel1/gaef-user-service/store"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	// read command-line args
	var prod bool
	flag.BoolVar(&prod, "production", false, "indicates the service is used for production")
	flag.Parse()

	// read environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var jwtSecret string
	if prod {
		gin.SetMode(gin.ReleaseMode)
		jwtSecret = os.Getenv("JWT_SECRET")
	} else {
		jwtSecret = "debug-jwt-secret"
	}
	port := os.Getenv("PORT")
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
	str := store.NewMongoStore(client.Database(dbName).Collection(collectionName))
	phv := domain.PasswordHasherVerifier{}
	svc := service.New(phv, phv, str, str, str, str, str)
	hdl := handler.New(svc, svc, svc, svc, svc, []byte(jwtSecret))
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
	r.Run(fmt.Sprintf("0.0.0.0:%s", port))
}
