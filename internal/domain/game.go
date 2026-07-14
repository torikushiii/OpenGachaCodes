package domain

type Game struct {
	Slug string `json:"slug" bson:"slug"`
	Name string `json:"name" bson:"name"`
}
