package group_test

import (
	"context"
	"github.com/gabrielseibel1/gaef/client/group"
	"testing"

	"github.com/gabrielseibel1/gaef/client/domain"
	"github.com/gabrielseibel1/gaef/client/user"
)

func TestClient_CRUD_Localhost8081(t *testing.T) {
	// we need a users client because groups API has authentication

	ctx := context.TODO()

	usersClient := user.Client{URL: "http://localhost:8080/api/v0/users/"}
	groupsClient := group.Client{URL: "http://localhost:8081/api/v0/groups/"}

	// create group with three users
	user1ID, err := usersClient.SignUp(ctx, "1", "1@gmail.com", "test1231")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	user2ID, err := usersClient.SignUp(ctx, "2", "2@gmail.com", "test1232")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	user3ID, err := usersClient.SignUp(ctx, "3", "3@gmail.com", "test1233")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	token, err := usersClient.Login(ctx, "1@gmail.com", "test1231")
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}
	g := domain.Group{
		Name:        "G",
		PictureURL:  "example.com",
		Description: "Gg",
		Members: []domain.User{
			{
				ID:   user1ID,
				Name: "A",
			},
			{
				ID:   user2ID,
				Name: "B",
			},
			{
				ID:   user3ID,
				Name: "C",
			},
		},
		Leaders: []domain.User{
			{
				ID:   user1ID,
				Name: "A",
			},
			{
				ID:   user3ID,
				Name: "C",
			},
		},
	}
	g, err = groupsClient.CreateGroup(ctx, token, g)
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}

	// create another group for the collections queries
	// to have multiple elements in the results
	g.Name = "H"
	g.Description = "Hh"
	g, err = groupsClient.CreateGroup(ctx, token, g)
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}

	// participating groups
	groups, err := groupsClient.ParticipatingGroups(ctx, token)
	if err != nil {
		t.Fatalf("groupsClient.ParticipatingGroups() = err: %s", err.Error())
	}
	if len(groups) != 2 {
		t.Fatalf("expected two groups, but groupsClient.ParticipatingGroups() = %v", groups)
	}

	// leading groups
	groups, err = groupsClient.LeadingGroups(ctx, token)
	if err != nil {
		t.Fatalf("groupsClient.LeadingGroups() = err: %s", err.Error())
	}
	if len(groups) != 2 {
		t.Fatalf("expected two groups, but groupsClient.LeadingGroups() = %v", groups)
	}

	// read group
	g, err = groupsClient.ReadGroup(ctx, token, groups[0].ID)
	if err != nil {
		t.Fatalf("groupsClient.ReadGroup() = err: %s", err.Error())
	}

	// read leading groups
	g, err = groupsClient.ReadLeadingGroup(ctx, token, groups[0].ID)
	if err != nil {
		t.Fatalf("groupsClient.ReadLeadingGroup() = err: %s", err.Error())
	}
	isLeader, err := groupsClient.IsGroupLeader(ctx, token, groups[1].ID)
	if err != nil {
		t.Fatalf("groupsClient.IsGroupLeader() = err: %s", err.Error())
	}
	if !isLeader {
		t.Fatalf("groupsClient.IsGroupLeader() = bool: %v", isLeader)
	}

	// update group
	g.Name = "I"
	g.Description = "Ii"
	g.Members, g.Leaders = g.Leaders, g.Members
	g, err = groupsClient.UpdateGroup(ctx, token, g)
	if err != nil {
		t.Fatalf("groupsClient.UpdateGroup() = err: %s", err.Error())
	}

	// delete group
	_, err = groupsClient.DeleteGroup(ctx, token, groups[0].ID)
	if err != nil {
		t.Fatalf("groupsClient.DeleteGroup() = err: %s", err.Error())
	}
}
