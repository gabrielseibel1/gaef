package store

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gabrielseibel1/gaef/encounter-proposal/domain"
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

func (m Mongo) Create(ctx context.Context, ep domain.EncounterProposal) (domain.EncounterProposal, error) {
	ep.ID = ""
	result, err := m.collection.InsertOne(ctx, ep)
	if err != nil {
		return domain.EncounterProposal{}, err
	}
	ep.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return ep, nil
}

func (m Mongo) ReadPaged(ctx context.Context, page int) ([]domain.EncounterProposal, error) {
	opts := options.Find().SetSort(bson.M{"_id": 1}).SetSkip(int64(page * pageSize))
	cursor, err := m.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}

	var eps []domain.EncounterProposal
	if err := cursor.All(ctx, &eps); err != nil {
		return nil, err
	}
	return eps, nil
}

func (m Mongo) ReadByUser(ctx context.Context, id string) ([]domain.EncounterProposal, error) {
	panic("not implemented")
}

func (m Mongo) ReadByID(ctx context.Context, id string) (domain.EncounterProposal, error) {
	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.EncounterProposal{}, err
	}

	result := m.collection.FindOne(ctx, bson.M{"_id": hex})
	if result.Err() != nil {
		return domain.EncounterProposal{}, result.Err()
	}

	var ep domain.EncounterProposal
	if err = result.Decode(&ep); err != nil {
		return domain.EncounterProposal{}, err
	}
	return ep, nil
}

func (m Mongo) Update(ctx context.Context, ep domain.EncounterProposal) (domain.EncounterProposal, error) {
	hex, err := primitive.ObjectIDFromHex(ep.ID)
	if err != nil {
		return domain.EncounterProposal{}, err
	}

	ep.ID = ""
	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": hex}, bson.M{"$set": ep})
	if err != nil {
		return domain.EncounterProposal{}, err
	}
	if result.ModifiedCount != 1 {
		return domain.EncounterProposal{}, errors.New("no such encounter proposal")
	}

	ep.ID = hex.Hex()
	return ep, nil
}

func (m Mongo) Delete(ctx context.Context, id string) error {
	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := m.collection.DeleteOne(ctx, bson.M{"_id": hex})
	if err != nil {
		return err
	}
	if result.DeletedCount != 1 {
		return errors.New("no such encounter proposal")
	}

	return nil
}

// Append TODO: maybe don't return the object
func (m Mongo) Append(ctx context.Context, epID string, app domain.Application) (domain.EncounterProposal, error) {
	hex, err := primitive.ObjectIDFromHex(epID)
	if err != nil {
		return domain.EncounterProposal{}, err
	}

	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": hex}, bson.M{"$push": bson.M{"applications": app}})
	if err != nil {
		return domain.EncounterProposal{}, err
	}
	if result.ModifiedCount != 1 {
		return domain.EncounterProposal{}, errors.New("no such encounter proposal")
	}

	return m.ReadByID(ctx, epID)
}

func (m Mongo) IsGroupLeader(ctx context.Context, groupID string, userID string) (bool, error) {
	panic("not implemented")
}

var pageSize = 50
