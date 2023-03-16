package group

import (
	"github.com/gabrielseibel1/client/user"
	"net/http"
	"testing"

	"github.com/gabrielseibel1/client/domain"
)

func TestClient_CRUD_Localhost8081(t *testing.T) {
	// we need a users client because groups API has authentication
	usersClient := user.Client{
		Host:       "localhost:8080",
		BasePath:   "/api/v0/users/",
		HTTPClient: http.Client{},
	}

	groupsClient := Client{
		Host:       "localhost:8081",
		BasePath:   "/api/v0/groups/",
		HTTPClient: http.Client{},
	}

	// create group with three users
	user1ID, err := usersClient.SignUp("A", "a@gmail.com", "test123a")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	user2ID, err := usersClient.SignUp("B", "b@gmail.com", "test123b")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	user3ID, err := usersClient.SignUp("C", "c@gmail.com", "test123c")
	if err != nil {
		t.Fatalf("usersClient.SignUp = err: %s", err.Error())
	}
	token, err := usersClient.Login("a@gmail.com", "test123a")
	if err != nil {
		t.Fatalf("usersClient.Login = err: %s", err.Error())
	}
	groupsClient.Token = token
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
	g, err = groupsClient.CreateGroup(g)
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}

	// create another group for the collections queries
	// to have multiple elements in the results
	g.Name = "H"
	g.Description = "Hh"
	g, err = groupsClient.CreateGroup(g)
	if err != nil {
		t.Fatalf("groupsClient.CreateGroup = err: %s", err.Error())
	}

	// participating groups
	groups, err := groupsClient.ParticipatingGroups()
	if err != nil {
		t.Fatalf("groupsClient.ParticipatingGroups() = err: %s", err.Error())
	}
	if len(groups) != 2 {
		t.Fatalf("expected two groups, but groupsClient.ParticipatingGroups() = %v", groups)
	}

	// leading groups
	groups, err = groupsClient.LeadingGroups()
	if err != nil {
		t.Fatalf("groupsClient.LeadingGroups() = err: %s", err.Error())
	}
	if len(groups) != 2 {
		t.Fatalf("expected two groups, but groupsClient.LeadingGroups() = %v", groups)
	}

	// read group
	g, err = groupsClient.ReadGroup(groups[0].ID)
	if err != nil {
		t.Fatalf("groupsClient.ReadGroup() = err: %s", err.Error())
	}

	// read leading groups
	g, err = groupsClient.ReadLeadingGroup(groups[0].ID)
	if err != nil {
		t.Fatalf("groupsClient.ReadLeadingGroup() = err: %s", err.Error())
	}
	g, err = groupsClient.ReadLeadingGroup(groups[1].ID)
	if err != nil {
		t.Fatalf("groupsClient.ReadLeadingGroup() = err: %s", err.Error())
	}

	// update group
	g.Name = "I"
	g.Description = "Ii"
	g.Members, g.Leaders = g.Leaders, g.Members
	g, err = groupsClient.UpdateGroup(g)
	if err != nil {
		t.Fatalf("groupsClient.UpdateGroup() = err: %s", err.Error())
	}

	// delete group
	_, err = groupsClient.DeleteGroup(groups[0].ID)
	if err != nil {
		t.Fatalf("groupsClient.DeleteGroup() = err: %s", err.Error())
	}
}
