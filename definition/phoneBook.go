package definition

import (
	"net/url"
)

type IPhoneBook interface {
	GetContactWithPagination(pageParam []string) ([]*Contact, error)
	AddContact(contact *Contact) (string, error)
	UpdateContact(id string, updatedContact *Contact) (int64, error)
	DeleteContact(id string) (int64, error)
	SearchContact(query url.Values) ([]*Contact, error)
}
