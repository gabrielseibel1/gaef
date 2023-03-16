package domain

type Group struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	PictureURL  string `json:"pictureUrl,omitempty"`
	Description string `json:"description,omitempty"`
	Members     []User `json:"members,omitempty"`
	Leaders     []User `json:"leaders,omitempty"`
}
