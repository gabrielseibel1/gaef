package store

import (
	"context"
	"errors"

	"github.com/gabrielseibel1/gaef/user/domain"
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

func (ms MongoStore) Create(user *domain.UserWithHashedPassword, ctx context.Context) (string, error) {
	res, err := ms.collection.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}
	id := res.InsertedID.(primitive.ObjectID).Hex()
	return id, nil
}

func (ms MongoStore) ReadByID(id string, ctx context.Context) (*domain.User, error) {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	res := ms.collection.FindOne(ctx, bson.M{"_id": hexID})
	if res.Err() != nil {
		return nil, res.Err()
	}
	var user domain.UserWithHashedPassword
	err = res.Decode(&user)
	if err != nil {
		return nil, err
	}
	user.ID = id
	return domain.ToSimplifiedUser(&user), err
}

func (ms MongoStore) ReadSensitiveByEmail(email string, ctx context.Context) (*domain.UserWithHashedPassword, error) {
	res := ms.collection.FindOne(ctx, bson.M{"email": email})
	if res.Err() != nil {
		return nil, res.Err()
	}
	var user domain.UserWithHashedPassword
	err := res.Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (ms MongoStore) Update(user *domain.User, ctx context.Context) (*domain.User, error) {
	hexID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return nil, err
	}
	res, err := ms.collection.UpdateOne(ctx, bson.M{"_id": hexID}, bson.M{"$set": user})
	if err != nil {
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, errors.New("no such user")
	}
	return user, nil
}

func (ms MongoStore) Delete(id string, ctx context.Context) error {
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
