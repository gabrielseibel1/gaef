package store

import (
	"context"
	"errors"
	"gaef-group-service/domain"

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

func (s MongoStore) CreateGroup(ctx context.Context, group domain.Group) (domain.Group, error) {
	res, err := s.collection.InsertOne(ctx, group)
	if err != nil {
		return domain.Group{}, err
	}
	id := res.InsertedID.(primitive.ObjectID).Hex()
	group.ID = id
	return group, nil
}

func (s MongoStore) ReadGroups(ctx context.Context, userID string) ([]domain.Group, error) {
	// hexID, err := primitive.ObjectIDFromHex(userID)
	// if err != nil {
	// return nil, err
	// }

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

	var groups []domain.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (s MongoStore) ReadGroup(ctx context.Context, id string) (domain.Group, error) {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.Group{}, err
	}

	res := s.collection.FindOne(ctx, bson.M{"_id": hexID})
	if res.Err() != nil {
		return domain.Group{}, res.Err()
	}

	var group domain.Group
	err = res.Decode(&group)
	if err != nil {
		return domain.Group{}, err
	}
	return group, err
}

func (s MongoStore) UpdateGroup(ctx context.Context, group domain.Group) (domain.Group, error) {
	hexID, err := primitive.ObjectIDFromHex(group.ID)
	if err != nil {
		return domain.Group{}, err
	}
	group.ID = "" // so that mongo doesn't think we are updating the id
	res, err := s.collection.UpdateOne(ctx, bson.M{"_id": hexID}, bson.M{"$set": group})
	if err != nil {
		return domain.Group{}, err
	}
	if res.MatchedCount == 0 {
		return domain.Group{}, errors.New("no such group")
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
