package definition

import "go.mongodb.org/mongo-driver/bson/primitive"

type Contact struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName  string             `json:"lastName,omitempty" bson:"lastName,omitempty"`
	Phone     string             `json:"phone,omitempty" bson:"phone,omitempty"`
	Address   string             `json:"address,omitempty" bson:"address,omitempty"`
}
