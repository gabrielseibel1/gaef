package encounterProposal_test

import (
	"context"
	"fmt"
	"github.com/gabrielseibel1/gaef/client/encounter-proposal"
	"github.com/gabrielseibel1/gaef/client/group"
	"github.com/gabrielseibel1/gaef/client/user"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func testWithURLs(t *testing.T, userServiceURL, groupServiceURL, encounterProposalServiceURL string) {
	ctx := context.TODO()

	usersClient := user.Client{URL: userServiceURL}
	groupsClient := group.Client{URL: groupServiceURL}
	encounterProposalClient := encounterProposal.Client{URL: encounterProposalServiceURL}

	// health check
	err := usersClient.Health(ctx)
	if err != nil {
		t.Fatalf("usersClient.Health = err: %s", err.Error())
	}
	err = groupsClient.Health(ctx)
	if err != nil {
		t.Fatalf("groupsClient.Health = err: %s", err.Error())
	}
	err = encounterProposalClient.Health(ctx)
	if err != nil {
		t.Fatalf("encounterProposalClient.Health = err: %s", err.Error())
	}

	// create two users and three groups:
	// user 1 - leads 1 group (g1)
	// user 2 - leads 2 groups (g2, g3)
	user1ID, err := usersClient.SignUp(ctx, types.User{Name: "1", Email: "eptest_1@gmail.com"}, "test1231")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	user2ID, err := usersClient.SignUp(ctx, types.User{Name: "2", Email: "eptest_2@gmail.com"}, "test1232")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	token1, err := usersClient.Login(ctx, "eptest_1@gmail.com", "test1231")
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}
	token2, err := usersClient.Login(ctx, "eptest_2@gmail.com", "test1232")
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}
	g1, err := groupsClient.CreateGroup(ctx, token1, types.Group{
		Name:        "G",
		PictureURL:  "example.com",
		Description: "Gg",
		Members:     []types.User{{ID: user1ID, Name: "1"}},
		Leaders:     []types.User{{ID: user1ID, Name: "1"}},
	})
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}
	g2, err := groupsClient.CreateGroup(ctx, token2, types.Group{
		Name:        "H",
		PictureURL:  "example.com",
		Description: "Hh",
		Members:     []types.User{{ID: user2ID, Name: "2"}},
		Leaders:     []types.User{{ID: user2ID, Name: "2"}},
	})
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}
	g3, err := groupsClient.CreateGroup(ctx, token2, types.Group{
		Name:        "I",
		PictureURL:  "example.com",
		Description: "Ii",
		Members:     []types.User{{ID: user2ID, Name: "2"}},
		Leaders:     []types.User{{ID: user2ID, Name: "2"}},
	})
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}
	defer func(usersClient user.Client, groupsClient group.Client, ctx context.Context, token, id string) {
		_, err := usersClient.DeleteUser(ctx, token, id)
		assert.Nil(t, err)
		_, err = usersClient.DeleteUser(ctx, token2, user2ID)
		assert.Nil(t, err)
		_, err = groupsClient.DeleteGroup(ctx, token1, g1.ID)
		assert.Nil(t, err)
		_, err = groupsClient.DeleteGroup(ctx, token2, g2.ID)
		assert.Nil(t, err)
		_, err = groupsClient.DeleteGroup(ctx, token2, g3.ID)
		assert.Nil(t, err)
	}(usersClient, groupsClient, ctx, token1, user1ID)

	// create three EPs:
	// 1 - created by g2 (leader user 2)
	// 2 - created by g3 (leader user 2)
	// 3 - created by g1 (leader user 1)
	createdEP1ID, err := encounterProposalClient.CreateEP(ctx, token2, types.EncounterProposal{
		EncounterSpecification: types.EncounterSpecification{
			Name:        "EP1",
			Description: "ep1 desc",
			Time:        time.Now().Add(time.Hour * 24),
		},
		Creator: g2,
	})
	if err != nil {
		t.Fatalf("encounterProposalClient.CreateEP() = err: %s", err.Error())
	}
	createdEP2ID, err := encounterProposalClient.CreateEP(ctx, token2, types.EncounterProposal{
		EncounterSpecification: types.EncounterSpecification{
			Name:        "EP2",
			Description: "ep2 desc",
			Time:        time.Now().Add(time.Hour * 24),
		},
		Creator: g3,
	})
	if err != nil {
		t.Fatalf("encounterProposalClient.CreateEP() = err: %s", err.Error())
	}
	createdEP3ID, err := encounterProposalClient.CreateEP(ctx, token1, types.EncounterProposal{
		EncounterSpecification: types.EncounterSpecification{
			Name:        "EP3",
			Description: "ep3 desc",
			Time:        time.Now().Add(time.Hour * 24),
		},
		Creator: g1,
	})
	if err != nil {
		t.Fatalf("encounterProposalClient.CreateEP() = err: %s", err.Error())
	}

	// assert "mine" collection has only first two created EPs
	mine, err := encounterProposalClient.Mine(ctx, token2)
	if err != nil {
		t.Fatalf("encounterProposalClient.Mine() = err: %s", err.Error())
	}
	if len(mine) != 2 {
		t.Fatalf("got len == %d, want len == 2", len(mine))
	}
	if mine[0].ID != createdEP1ID {
		t.Fatalf("got %v, want %v", mine[0].ID, createdEP1ID)
	}
	if mine[1].ID != createdEP2ID {
		t.Fatalf("got %v, want %v", mine[1].ID, createdEP2ID)
	}

	// assert page 0 has all three EPs (3 is most likely below page size that defaults to 50)
	page0, err := encounterProposalClient.Page(ctx, token2, 0)
	if err != nil {
		t.Fatalf("encounterProposalClient.Page() = err: %s", err.Error())
	}
	if len(page0) != 3 {
		t.Fatalf("got len == %d, want len == 3", len(page0))
	}
	if page0[0].ID != createdEP1ID {
		t.Fatalf("got %v, want %v", page0[0].ID, createdEP1ID)
	}
	if page0[1].ID != createdEP2ID {
		t.Fatalf("got %v, want %v", page0[1].ID, createdEP2ID)
	}
	if page0[2].ID != createdEP3ID {
		t.Fatalf("got %v, want %v", page0[2].ID, createdEP3ID)
	}

	// assert that next page has no content
	page1, err := encounterProposalClient.Page(ctx, token2, 1)
	if err != nil {
		t.Fatalf("encounterProposalClient.Page() = err: %s", err.Error())
	}
	if len(page1) != 0 {
		t.Fatalf("got len == %d, want len == 0", len(page1))
	}

	// read each EP and check equals to created one
	readEP1, err := encounterProposalClient.ReadEP(ctx, token2, createdEP1ID)
	if err != nil {
		t.Fatalf("encounterProposalClient.ReadEP() = err: %s", err.Error())
	}
	if readEP1.ID != createdEP1ID {
		t.Fatalf("got %v, want %v", readEP1.ID, createdEP1ID)
	}
	readEP2, err := encounterProposalClient.ReadEP(ctx, token2, createdEP2ID)
	if err != nil {
		t.Fatalf("encounterProposalClient.ReadEP() = err: %s", err.Error())
	}
	if readEP2.ID != createdEP2ID {
		t.Fatalf("got %v, want %v", readEP2.ID, createdEP2ID)
	}
	readEP3, err := encounterProposalClient.ReadEP(ctx, token2, createdEP3ID)
	if err != nil {
		t.Fatalf("encounterProposalClient.ReadEP() = err: %s", err.Error())
	}
	if readEP3.ID != createdEP3ID {
		t.Fatalf("got %v, want %v", readEP3.ID, createdEP3ID)
	}

	// update an encounter proposal and assert changes were made
	readEP1.Name = "Updated Name"
	updatedEP1, err := encounterProposalClient.UpdateEP(ctx, token2, readEP1)
	if err != nil {
		t.Fatalf("encounterProposalClient.UpdateEP() = err: %s", err.Error())
	}
	if got, want := updatedEP1, readEP1; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	// append an application to an encounter proposal (use token 1, user 1, leader of g1)
	appliedMessage, err := encounterProposalClient.ApplyToEP(ctx, token1, readEP2.ID, types.Application{
		Description: "application1",
		Applicant:   g1,
	})
	if err != nil {
		t.Fatalf("encounterProposalClient.ApplyToEP() = err: %s", err.Error())
	}
	if got, want := appliedMessage, "applied for encounter proposal "+readEP2.ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	// delete application just appended
	deletedAppMessage, err := encounterProposalClient.DeleteApplication(ctx, token2, readEP2.ID, g1.ID)
	if err != nil {
		t.Fatalf("encounterProposalClient.DeleteApplication() = err: %s", err.Error())
	}
	if got, want := deletedAppMessage, fmt.Sprintf("deleted application %s of encounter proposal %s", g1.ID, readEP2.ID); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	// delete all encounter proposals
	deletedMessage, err := encounterProposalClient.DeleteEP(ctx, token2, createdEP1ID)
	if err != nil {
		t.Fatalf("encounterProposalClient.DeleteEP() = err: %s", err.Error())
	}
	if got, want := deletedMessage, "deleted encounter proposal "+createdEP1ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	deletedMessage, err = encounterProposalClient.DeleteEP(ctx, token2, createdEP2ID)
	if err != nil {
		t.Fatalf("encounterProposalClient.DeleteEP() = err: %s", err.Error())
	}
	if got, want := deletedMessage, "deleted encounter proposal "+createdEP2ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	deletedMessage, err = encounterProposalClient.DeleteEP(ctx, token1, createdEP3ID)
	if err != nil {
		t.Fatalf("encounterProposalClient.DeleteEP() = err: %s", err.Error())
	}
	if got, want := deletedMessage, "deleted encounter proposal "+createdEP3ID; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestClient_Localhost8082(t *testing.T) {
	testWithURLs(
		t,
		"http://localhost:8080/api/v0/users/",
		"http://localhost:8081/api/v0/groups/",
		"http://localhost:8082/api/v0/encounter-proposals/",
	)
}

func TestClient_Production(t *testing.T) {
	testWithURLs(
		t,
		"https://gaef-user-service.onrender.com/api/v0/users/",
		"https://gaef-group-service.onrender.com/api/v0/groups/",
		"https://gaef-encounter-proposal-service.onrender.com/api/v0/encounter-proposals/",
	)
}
