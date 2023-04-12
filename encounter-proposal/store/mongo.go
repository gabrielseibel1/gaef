package store

import (
	"context"
	"errors"
	"github.com/gabrielseibel1/gaef/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

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

func (m Mongo) Create(ctx context.Context, ep types.EncounterProposal) (types.EncounterProposal, error) {
	ep.ID = ""
	// create with a non-nil slice of len 0 to be pushable
	ep.Applications = []types.Application{}
	result, err := m.collection.InsertOne(ctx, ep)
	if err != nil {
		return types.EncounterProposal{}, err
	}
	ep.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return ep, nil
}

func (m Mongo) ReadPaged(ctx context.Context, page int) ([]types.EncounterProposal, error) {
	opts := options.Find().SetSort(bson.M{"_id": 1}).SetSkip(int64(page * pageSize))
	cursor, err := m.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}

	var eps []types.EncounterProposal
	if err := cursor.All(ctx, &eps); err != nil {
		return nil, err
	}
	return eps, nil
}

func (m Mongo) ReadByID(ctx context.Context, id string) (types.EncounterProposal, error) {
	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.EncounterProposal{}, err
	}

	result := m.collection.FindOne(ctx, bson.M{"_id": hex})
	if result.Err() != nil {
		return types.EncounterProposal{}, result.Err()
	}

	var ep types.EncounterProposal
	if err = result.Decode(&ep); err != nil {
		return types.EncounterProposal{}, err
	}
	return ep, nil
}

func (m Mongo) ReadByGroupIDs(ctx context.Context, groupIDs []string) ([]types.EncounterProposal, error) {
	var ids bson.A
	for _, gid := range groupIDs {
		ids = append(ids, gid)
	}
	cursor, err := m.collection.Find(ctx, bson.M{"creator._id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}

	var eps []types.EncounterProposal
	if err := cursor.All(ctx, &eps); err != nil {
		return nil, err
	}
	return eps, nil
}

func (m Mongo) Update(ctx context.Context, ep types.EncounterProposal) (types.EncounterProposal, error) {
	hex, err := primitive.ObjectIDFromHex(ep.ID)
	if err != nil {
		return types.EncounterProposal{}, err
	}

	ep.ID = ""
	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": hex}, bson.M{"$set": ep})
	if err != nil {
		return types.EncounterProposal{}, err
	}
	if result.ModifiedCount != 1 {
		return types.EncounterProposal{}, errors.New("no such encounter proposal")
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

func (m Mongo) AppendApplication(ctx context.Context, epID string, app types.Application) error {
	hex, err := primitive.ObjectIDFromHex(epID)
	if err != nil {
		return err
	}

	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": hex}, bson.M{"$push": bson.M{"applications": app}})
	if err != nil {
		return err
	}
	if result.ModifiedCount != 1 {
		return errors.New("no such encounter proposal")
	}

	return nil
}

func (m Mongo) DeleteApplication(ctx context.Context, epID string, appID string) error {
	hex, err := primitive.ObjectIDFromHex(epID)
	if err != nil {
		return err
	}

	result, err := m.collection.UpdateOne(ctx, bson.M{"_id": hex}, bson.M{"$pull": bson.M{"applications": bson.M{"applicant._id": appID}}})
	if err != nil {
		return err
	}
	if result.ModifiedCount != 1 {
		return errors.New("no such encounter proposal")
	}

	return nil
}

var pageSize = 50
