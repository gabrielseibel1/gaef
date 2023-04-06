package group_test

import (
	"context"
	"github.com/gabrielseibel1/gaef/client/group"
	"github.com/gabrielseibel1/gaef/client/user"
	"github.com/gabrielseibel1/gaef/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testWithURLs(t *testing.T, userServiceURL, groupServiceURL string) {
	// we need a users client because groups API has authentication

	ctx := context.TODO()

	usersClient := user.Client{URL: userServiceURL}
	groupsClient := group.Client{URL: groupServiceURL}

	// create group with three users
	user1ID, err := usersClient.SignUp(ctx, types.User{Name: "1", Email: "grouptest1@gmail.com"}, "test1231")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	user2ID, err := usersClient.SignUp(ctx, types.User{Name: "2", Email: "grouptest2@gmail.com"}, "test1232")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	user3ID, err := usersClient.SignUp(ctx, types.User{Name: "3", Email: "grouptest3@gmail.com"}, "test1233")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	token1, err := usersClient.Login(ctx, "grouptest1@gmail.com", "test1231")
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}
	token2, err := usersClient.Login(ctx, "grouptest2@gmail.com", "test1232")
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}
	token3, err := usersClient.Login(ctx, "grouptest3@gmail.com", "test1233")
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}
	// cleanup afterward
	defer func(usersClient user.Client, ctx context.Context, token, id string) {
		_, err := usersClient.DeleteUser(ctx, token, id)
		assert.Nil(t, err)
		_, err = usersClient.DeleteUser(ctx, token2, user2ID)
		assert.Nil(t, err)
		_, err = usersClient.DeleteUser(ctx, token3, user3ID)
		assert.Nil(t, err)
	}(usersClient, ctx, token1, user1ID)

	createdGroup1, err := groupsClient.CreateGroup(ctx, token1, types.Group{
		Name:        "G",
		PictureURL:  "example.com",
		Description: "Gg",
		Members:     []types.User{{ID: user1ID, Name: "A"}, {ID: user2ID, Name: "B"}, {ID: user3ID, Name: "C"}},
		Leaders:     []types.User{{ID: user1ID, Name: "A"}, {ID: user3ID, Name: "C"}},
	})
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}

	// create another group for the collections queries
	// to have multiple elements in the results
	createdGroup2, err := groupsClient.CreateGroup(ctx, token1, types.Group{
		Name:        "H",
		PictureURL:  "example.com",
		Description: "Hh",
		Members:     []types.User{{ID: user1ID, Name: "A"}, {ID: user2ID, Name: "B"}, {ID: user3ID, Name: "C"}},
		Leaders:     []types.User{{ID: user1ID, Name: "A"}, {ID: user3ID, Name: "C"}},
	})
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}

	// participating groups
	participatingGroups, err := groupsClient.ParticipatingGroups(ctx, token1)
	if err != nil {
		t.Fatalf("groupsClient.ParticipatingGroups() = err: %s", err.Error())
	}
	if len(participatingGroups) != 2 {
		t.Fatalf("expected two groups, but groupsClient.ParticipatingGroups() = %v", participatingGroups)
	}

	// leading groups
	leadingGroups, err := groupsClient.LeadingGroups(ctx, token1)
	if err != nil {
		t.Fatalf("groupsClient.LeadingGroups() = err: %s", err.Error())
	}
	if len(leadingGroups) != 2 {
		t.Fatalf("expected two groups, but groupsClient.LeadingGroups() = %v", leadingGroups)
	}

	// read group
	_, err = groupsClient.ReadGroup(ctx, token1, leadingGroups[0].ID)
	if err != nil {
		t.Fatalf("groupsClient.ReadGroup() = err: %s", err.Error())
	}

	// read leading groups
	_, err = groupsClient.ReadLeadingGroup(ctx, token1, leadingGroups[0].ID)
	if err != nil {
		t.Fatalf("groupsClient.ReadLeadingGroup() = err: %s", err.Error())
	}
	isLeader, err := groupsClient.IsGroupLeader(ctx, token1, leadingGroups[1].ID)
	if err != nil {
		t.Fatalf("groupsClient.IsGroupLeader() = err: %s", err.Error())
	}
	if !isLeader {
		t.Fatalf("groupsClient.IsGroupLeader() = bool: %v", isLeader)
	}

	// update group
	createdGroup1.Name = "I"
	createdGroup1.Description = "Ii"
	createdGroup1.Members, createdGroup1.Leaders = createdGroup1.Leaders, createdGroup1.Members
	_, err = groupsClient.UpdateGroup(ctx, token1, createdGroup1)
	if err != nil {
		t.Fatalf("groupsClient.UpdateGroup() = err: %s", err.Error())
	}

	// delete groups
	_, err = groupsClient.DeleteGroup(ctx, token1, createdGroup1.ID)
	if err != nil {
		t.Fatalf("groupsClient.DeleteGroup() = err: %s", err.Error())
	}
	_, err = groupsClient.DeleteGroup(ctx, token1, createdGroup2.ID)
	if err != nil {
		t.Fatalf("groupsClient.DeleteGroup() = err: %s", err.Error())
	}
}

func TestClient_Localhost8081(t *testing.T) {
	testWithURLs(
		t,
		"http://localhost:8080/api/v0/users/",
		"http://localhost:8081/api/v0/groups/",
	)
}

func TestClient_Production(t *testing.T) {
	testWithURLs(
		t,
		"https://gaef-user-service.onrender.com/api/v0/users/",
		"https://gaef-group-service.onrender.com/api/v0/groups/",
	)
}
