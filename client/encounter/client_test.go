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

func testWithURLs(t *testing.T, userServiceURL, groupServiceURL, encounterServiceURL string) {
	ctx := context.TODO()

	usersClient := user.Client{URL: userServiceURL}
	groupsClient := group.Client{URL: groupServiceURL}
	encountersClient := encounter.Client{URL: encounterServiceURL}

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
	// cleanup afterward
	defer func(usersClient user.Client, ctx context.Context, token, id string) {
		_, err := usersClient.DeleteUser(ctx, token, id)
		assert.Nil(t, err)
		_, err = groupsClient.DeleteGroup(ctx, token1, g1.ID)
		assert.Nil(t, err)
	}(usersClient, ctx, token1, user1ID)

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

func TestClient_Localhost8083(t *testing.T) {
	testWithURLs(
		t,
		"http://localhost:8080/api/v0/users/",
		"http://localhost:8081/api/v0/groups/",
		"http://localhost:8083/api/v0/encounters/",
	)
}

func TestClient_Production(t *testing.T) {
	testWithURLs(
		t,
		"https://gaef-user-service.onrender.com/api/v0/users/",
		"https://gaef-group-service.onrender.com/api/v0/groups/",
		"https://gaef-encounter-service.onrender.com/api/v0/encounters/",
	)
}
