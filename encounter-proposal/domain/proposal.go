package domain

import "time"

type EncounterProposal struct {
	ID           string `bson:"_id,omitempty"`
	Name         string
	Description  string
	Time         time.Time
	Creator      Group
	Applications []Application
}

type Group struct {
	ID   string
	Name string
}

type Application struct {
	Description string
	Creator     Group
}
