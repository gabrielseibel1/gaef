package domain

type User struct {
	ID   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

type Group struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"name"`
	PictureURL  string `json:"pictureUrl" bson:"pictureUrl"`
	Description string `json:"description" bson:"description"`
	Members     []User `json:"members" bson:"members"`
	Leaders     []User `json:"leaders" bson:"leaders"`
}
