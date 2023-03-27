package types

import (
	"time"
)

type User struct {
	ID             string `json:"id" bson:"_id,omitempty"`
	Name           string `json:"name" bson:"name"`
	Email          string `json:"email" bson:"email"`
	Password       string `json:"password" bson:"password"`
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
	ID           string        `json:"id" bson:"_id,omitempty"`
	Name         string        `json:"name" bson:"name"`
	Description  string        `json:"description" bson:"description"`
	Time         time.Time     `json:"time" bson:"time"`
	Creator      Group         `json:"creator" bson:"creator"`
	Applications []Application `json:"applications" bson:"applications"`
}

type Application struct {
	Description string `json:"description" bson:"description"`
	Creator     Group  `json:"creator" bson:"creator"`
}
