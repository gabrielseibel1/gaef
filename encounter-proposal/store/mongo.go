package store

import (
	"context"

	"github.com/gabrielseibel1/gaef-encounter-proposal-service/domain"
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
	panic("not implemented")
}

func (m Mongo) ReadPaged(ctx context.Context, page int) ([]domain.EncounterProposal, error) {
	panic("not implemented")
}

func (m Mongo) ReadByUser(ctx context.Context, id string) ([]domain.EncounterProposal, error) {
	panic("not implemented")
}

func (m Mongo) ReadByID(ctx context.Context, id string) (domain.EncounterProposal, error) {
	panic("not implemented")
}

func (m Mongo) Update(ctx context.Context, ep domain.EncounterProposal) (domain.EncounterProposal, error) {
	panic("not implemented")
}

func (m Mongo) Delete(ctx context.Context, id string) error {
	panic("not implemented")
}

func (m Mongo) Append(ctx context.Context, epID string, app domain.Application) (domain.EncounterProposal, error) {
	panic("not implemented")
}

func (m Mongo) IsGroupLeader(ctx context.Context, groupID string, userID string) (bool, error) {
	panic("not implemented")
}
