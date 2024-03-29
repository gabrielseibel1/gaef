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

func New(collection *mongo.Collection) *MongoStore {
	return &MongoStore{
		collection: collection,
	}
}

func (s MongoStore) IsLeader(ctx context.Context, userID string, groupID string) (bool, error) {
	group, err := s.ReadGroup(ctx, groupID)
	if err != nil {
		return false, err
	}

	for _, leader := range group.Leaders {
		if leader.ID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (s MongoStore) CreateGroup(ctx context.Context, group types.Group) (types.Group, error) {
	group.ID = ""
	res, err := s.collection.InsertOne(ctx, group)
	if err != nil {
		return types.Group{}, err
	}
	id := res.InsertedID.(primitive.ObjectID).Hex()
	group.ID = id
	return group, nil
}

func (s MongoStore) ReadParticipatingGroups(ctx context.Context, userID string) ([]types.Group, error) {
	cursor, err := s.collection.Find(ctx, bson.M{
		"members": bson.M{
			"$elemMatch": bson.M{
				"_id": userID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []types.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (s MongoStore) ReadLeadingGroups(ctx context.Context, userID string) ([]types.Group, error) {
	cursor, err := s.collection.Find(ctx, bson.M{
		"leaders": bson.M{
			"$elemMatch": bson.M{
				"_id": userID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []types.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (s MongoStore) ReadGroup(ctx context.Context, id string) (types.Group, error) {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.Group{}, err
	}

	res := s.collection.FindOne(ctx, bson.M{"_id": hexID})
	if res.Err() != nil {
		return types.Group{}, res.Err()
	}

	var group types.Group
	err = res.Decode(&group)
	if err != nil {
		return types.Group{}, err
	}
	return group, err
}

func (s MongoStore) UpdateGroup(ctx context.Context, group types.Group) (types.Group, error) {
	hexID, err := primitive.ObjectIDFromHex(group.ID)
	if err != nil {
		return types.Group{}, err
	}
	group.ID = "" // so that mongo doesn't think we are updating the id
	res, err := s.collection.UpdateOne(ctx, bson.M{"_id": hexID}, bson.M{"$set": group})
	if err != nil {
		return types.Group{}, err
	}
	if res.MatchedCount == 0 {
		return types.Group{}, errors.New("no such group")
	}
	group.ID = hexID.Hex()
	return group, nil
}

func (s MongoStore) DeleteGroup(ctx context.Context, id string) error {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	res, err := s.collection.DeleteOne(ctx, bson.M{"_id": hexID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("no such group")
	}
	return nil
}
