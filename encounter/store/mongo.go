package store

import (
	"context"
	"errors"
	"github.com/gabrielseibel1/gaef/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo struct {
	collection *mongo.Collection // TODO: extract and depend on interface
}

func New(collection *mongo.Collection) Mongo {
	return Mongo{
		collection: collection,
	}
}

func (m Mongo) CreateEncounter(ctx context.Context, e types.Encounter) (string, error) {
	e.ID = ""                         // don't want any id specified before insertion
	e.ConfirmedUsers = []types.User{} // create with a non-nil slice of len 0 to be pushable
	result, err := m.collection.InsertOne(ctx, e)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (m Mongo) ReadEncounter(ctx context.Context, id string) (types.Encounter, error) {
	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.Encounter{}, err
	}

	result := m.collection.FindOne(ctx, bson.M{"_id": hex})
	if result.Err() != nil {
		return types.Encounter{}, err
	}

	var e types.Encounter
	if err := result.Decode(&e); err != nil {
		return types.Encounter{}, err
	}
	return e, nil
}

func (m Mongo) ReadUserEncounters(ctx context.Context, userID string) ([]types.Encounter, error) {
	cursor, err := m.collection.Find(ctx, bson.M{"invitedUsers._id": userID})
	if err != nil {
		return nil, err
	}

	var e []types.Encounter
	if err := cursor.All(ctx, &e); err != nil {
		return []types.Encounter{}, err
	}
	return e, nil
}

func (m Mongo) UpdateEncounter(ctx context.Context, e types.Encounter) (types.Encounter, error) {
	hex, err := primitive.ObjectIDFromHex(e.ID)
	if err != nil {
		return types.Encounter{}, err
	}

	e.ID = ""
	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": hex}, bson.M{"$set": e})
	if err != nil {
		return types.Encounter{}, err
	}
	if result.ModifiedCount != 1 {
		return types.Encounter{}, errors.New("no such encounter")
	}

	e.ID = hex.Hex()
	return e, nil
}

func (m Mongo) DeleteEncounter(ctx context.Context, id string) error {
	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := m.collection.DeleteOne(ctx, bson.M{"_id": hex})
	if err != nil {
		return err
	}
	if result.DeletedCount != 1 {
		return errors.New("no such encounter")
	}

	return nil
}

func (m Mongo) ConfirmEncounter(ctx context.Context, encID string, user types.User) error {
	hex, err := primitive.ObjectIDFromHex(encID)
	if err != nil {
		return err
	}

	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": hex}, bson.M{"$push": bson.M{"confirmedUsers": user}})
	if err != nil {
		return err
	}
	if result.ModifiedCount != 1 {
		return errors.New("no such encounter")
	}

	return nil
}

func (m Mongo) DeclineEncounter(ctx context.Context, encID, userID string) error {
	hex, err := primitive.ObjectIDFromHex(encID)
	if err != nil {
		return err
	}

	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": hex}, bson.M{"$pull": bson.M{"confirmedUsers": bson.M{"_id": userID}}})
	if err != nil {
		return err
	}
	if result.ModifiedCount != 1 {
		return errors.New("no such encounter")
	}

	return nil
}
