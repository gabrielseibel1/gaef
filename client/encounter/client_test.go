package encounter_test

import (
	"context"
	"github.com/gabrielseibel1/gaef/client/encounter"
	"github.com/gabrielseibel1/gaef/client/group"
	"github.com/gabrielseibel1/gaef/client/user"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClient_Localhost8083(t *testing.T) {
	ctx := context.TODO()

	usersClient := user.Client{URL: "http://localhost:8080/api/v0/users/"}
	groupsClient := group.Client{URL: "http://localhost:8081/api/v0/groups/"}
	encountersClient := encounter.Client{URL: "http://localhost:8083/api/v0/encounters/"}

	// create user and group
	user1ID, err := usersClient.SignUp(ctx, types.User{Name: "1", Email: "enctest_1@gmail.com"}, "test1231")
	assert.Nil(t, err)
	token1, err := usersClient.Login(ctx, "enctest_1@gmail.com", "test1231")
	assert.Nil(t, err)
	user1, err := usersClient.ReadUser(ctx, token1, user1ID)
	assert.Nil(t, err)
	g1, err := groupsClient.CreateGroup(ctx, token1, types.Group{
		Name:        "G",
		PictureURL:  "example.com",
		Description: "Gg",
		Members:     []types.User{{ID: user1ID, Name: "1"}},
		Leaders:     []types.User{{ID: user1ID, Name: "1"}},
	})
	assert.Nil(t, err)

	// define encounter
	enc1 := types.Encounter{
		EncounterSpecification: types.EncounterSpecification{
			Name:        "test-encounter-name-1",
			Description: "test-encounter-description-1",
			Location: types.Location{
				Name:      "test-encounter-location-1",
				Latitude:  42.87,
				Longitude: 87.42,
			},
			Time: time.Now().Round(time.Minute).UTC(), // without round and UTC we fail some assertions of equality
		},
		Groups:         []types.Group{g1},
		InvitedUsers:   []types.User{user1},
		ConfirmedUsers: []types.User{},
	}

	// create encounter
	enc1.ID, err = encountersClient.CreateEncounter(ctx, token1, enc1)
	assert.Nil(t, err)

	// query user encounters and verify first is the created one
	userEncounters, err := encountersClient.GetUserEncounters(ctx, token1)
	assert.Nil(t, err)
	assert.Equal(t, enc1, userEncounters[0])

	// query by id and verify equals definition
	readEncounter, err := encountersClient.GetEncounter(ctx, token1, enc1.ID)
	assert.Nil(t, err)
	assert.Equal(t, enc1, readEncounter)

	// modify encounter
	enc1.Name = "test-encounter-name-2"
	updatedEncounter, err := encountersClient.UpdateEncounter(ctx, token1, enc1)
	assert.Nil(t, err)
	assert.Equal(t, enc1, updatedEncounter)

	// confirm encounter
	confirmedID, err := encountersClient.ConfirmEncounter(ctx, token1, enc1.ID)
	assert.Nil(t, err)
	assert.Equal(t, enc1.ID, confirmedID)

	// decline encounter
	declinedID, err := encountersClient.DeclineEncounter(ctx, token1, enc1.ID)
	assert.Nil(t, err)
	assert.Equal(t, enc1.ID, declinedID)

	// delete encounter
	deletedID, err := encountersClient.DeleteEncounter(ctx, token1, enc1.ID)
	assert.Nil(t, err)
	assert.Equal(t, enc1.ID, deletedID)
}
