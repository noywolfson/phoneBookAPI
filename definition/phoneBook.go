package definition

import (
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
)

type IPhoneBook interface {
	GetContactWithPagination(pageParam []string) ([]*Contact, error)
	AddContact(contact *Contact) (*mongo.InsertOneResult, error)
	UpdateContact(id string, updatedContact *Contact) (int64, error)
	DeleteContact(id string) (int64, error)
	SearchContact(query url.Values) ([]*Contact, error)
}
