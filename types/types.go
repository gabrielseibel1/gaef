package types

import (
	"time"
)

type User struct {
	ID         string `json:"id" bson:"_id,omitempty"`
	Email      string `json:"email" bson:"email"`
	Name       string `json:"name" bson:"name"`
	PictureURL string `json:"pictureUrl" bson:"pictureUrl"`
}

type UserWithHashedPassword struct {
	ID             string `json:"id" bson:"_id,omitempty"`
	User           `json:"user" bson:"user"`
	HashedPassword string `json:"hashedPassword" bson:"hashedPassword"`
}

type Group struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"name"`
	PictureURL  string `json:"pictureUrl" bson:"pictureUrl"`
	Description string `json:"description" bson:"description"`
	Members     []User `json:"members" bson:"members"`
	Leaders     []User `json:"leaders" bson:"leaders"`
}

type EncounterProposal struct {
	ID                     string `json:"id" bson:"_id,omitempty"`
	EncounterSpecification `json:"encounterSpecification" bson:"encounterSpecification"`
	Creator                Group         `json:"creator" bson:"creator"`
	Applications           []Application `json:"applications" bson:"applications"`
}

type EncounterSpecification struct {
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	Location    Location  `json:"location" bson:"location"`
	Time        time.Time `json:"time" bson:"time"`
}

type Location struct {
	Name      string  `json:"name" bson:"name"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

type Application struct {
	Description string `json:"description" bson:"description"`
	Applicant   Group  `json:"applicant" bson:"applicant"`
}

type Encounter struct {
	ID                     string `json:"id" bson:"_id,omitempty"`
	EncounterSpecification `json:"encounterSpecification" bson:"encounterSpecification"`
	Groups                 []Group `json:"groups" bson:"groups"`
	InvitedUsers           []User  `json:"invitedUsers" bson:"invitedUsers"`
	ConfirmedUsers         []User  `json:"confirmedUsers" bson:"confirmedUsers"`
}
