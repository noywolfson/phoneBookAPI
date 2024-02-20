package definition

import (
	"net/url"
)

type IPhoneBook interface {
	GetContactWithPagination(pageParam []string) ([]*Contact, string, error)
	AddContact(contact *Contact) (string, string, error)
	UpdateContact(id string, updatedContact *Contact) (int64, string, error)
	DeleteContact(id string) (int64, string, error)
	SearchContact(query url.Values) ([]*Contact, string, error)
}
