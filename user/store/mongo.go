package store

import (
	"context"
	"errors"
	"github.com/gabrielseibel1/gaef/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoStore struct {
	collection *mongo.Collection
}

func NewMongoStore(collection *mongo.Collection) *MongoStore {
	return &MongoStore{
		collection: collection,
	}
}

func (ms MongoStore) Create(ctx context.Context, user types.UserWithHashedPassword) (string, error) {
	user.ID, user.User.ID = "", ""
	res, err := ms.collection.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}

	id := res.InsertedID.(primitive.ObjectID).Hex()
	user.User.ID = id
	err = ms.Update(ctx, user.User)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ms MongoStore) ReadByID(ctx context.Context, id string) (types.User, error) {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.User{}, err
	}

	res := ms.collection.FindOne(ctx, bson.M{"_id": hexID})
	if res.Err() != nil {
		return types.User{}, res.Err()
	}

	var user types.UserWithHashedPassword
	err = res.Decode(&user)
	if err != nil {
		return types.User{}, err
	}
	return user.User, err
}

func (ms MongoStore) ReadSensitiveByEmail(ctx context.Context, email string) (types.UserWithHashedPassword, error) {
	res := ms.collection.FindOne(ctx, bson.M{"user.email": email})
	if res.Err() != nil {
		return types.UserWithHashedPassword{}, res.Err()
	}

	var user types.UserWithHashedPassword
	err := res.Decode(&user)
	if err != nil {
		return types.UserWithHashedPassword{}, err
	}
	return user, err
}

func (ms MongoStore) Update(ctx context.Context, user types.User) error {
	hexID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return err
	}

	res, err := ms.collection.UpdateOne(ctx, bson.M{"_id": hexID}, bson.M{"$set": bson.M{"user": user}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("no such user")
	}
	return nil
}

func (ms MongoStore) Delete(ctx context.Context, id string) error {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	res, err := ms.collection.DeleteOne(ctx, bson.M{"_id": hexID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("no such user")
	}
	return nil
}
