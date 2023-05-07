package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gabrielseibel1/gaef/messenger"
	"github.com/gabrielseibel1/gaef/user/hasher"
	amqp "github.com/rabbitmq/amqp091-go"
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
	amqpURI := os.Getenv("AMQP_URI")
	amqpExchangeUpdates := os.Getenv("AMQP_EXCHANGE_UPDATES")
	amqpExchangeDeletes := os.Getenv("AMQP_EXCHANGE_DELETES")
	dbURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DATABASE")
	collectionName := os.Getenv("MONGODB_COLLECTION")

	// connect to mongoDB
	client, err := setupMongoDB(dbURI)
	if err != nil {
		log.Fatal(err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}(client, context.Background())

	// connect to rabbitmq
	channel, err := setupRabbitMQ(amqpURI, amqpExchangeUpdates, amqpExchangeDeletes)
	if err != nil {
		log.Fatal(err)
	}

	// instantiate and inject dependencies
	msg := messenger.New(json.Marshal, amqpExchangeUpdates, amqpExchangeDeletes, channel)
	str := store.NewMongoStore(client.Database(dbName).Collection(collectionName))
	phv := hasher.New()
	hdl := handler.New(phv, phv, str, str, str, str, str, msg, msg, []byte(jwtSecret))
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

func setupMongoDB(dbURI string) (*mongo.Client, error) {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(dbURI).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	return client, nil
}

func setupRabbitMQ(amqpURI, amqpExchangeUpdates, amqpExchangeDeletes string) (*amqp.Channel, error) {
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, err
	}
	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	err = channel.ExchangeDeclare(
		amqpExchangeUpdates,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = channel.ExchangeDeclare(
		amqpExchangeDeletes,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return channel, nil
}
