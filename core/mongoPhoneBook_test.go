package core

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"net/url"
	"phoneBook/config"
	"phoneBook/definition"
	"testing"
)

var (
	client               *mongo.Client
	ErrorNotExistContact = "invalid last name. name should include letters only"
	ErrorInvalidObjectID = "the provided hex string is not a valid ObjectID"
)

func TestAddContact(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	validContact := &definition.Contact{
		FirstName: "valid",
		LastName:  "contact",
		Phone:     "0545454524",
		Address:   "Tel Aviv",
	}
	validContactWithoutLastNameAndAddress := &definition.Contact{
		FirstName: "momo",
		Phone:     "0545454524",
	}
	invalidContactWithoutPhone := &definition.Contact{
		FirstName: "momo",
	}
	invalidContactPhone := &definition.Contact{
		FirstName: "valid",
		LastName:  "contact",
		Phone:     "054abc4524",
		Address:   "Tel Aviv",
	}
	invalidContactLastName := &definition.Contact{
		FirstName: "hi",
		LastName:  "cocot345",
		Phone:     "054abc4524",
		Address:   "Tel Aviv",
	}

	mt.Run("should add valid contact", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(validContact)
		assert.Nil(t, err)
	})

	mt.Run("should add contact without last name and address", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(validContactWithoutLastNameAndAddress)
		assert.Nil(t, err)
	})

	mt.Run("should not add contact without phone", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(invalidContactWithoutPhone)
		assert.EqualErrorf(t, err, ErrorMissingPhone, "Error should be: %v, got: %v", ErrorMissingPhone, err)
	})

	mt.Run("should not add contact with invalid phone", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(invalidContactPhone)
		assert.EqualErrorf(t, err, ErrorInvalidPhone, "Error should be: %v, got: %v", ErrorInvalidPhone, err)
	})

	mt.Run("should not add contact with invalid name", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(invalidContactLastName)
		assert.EqualErrorf(t, err, ErrorInvalidLastName, "Error should be: %v, got: %v", ErrorInvalidLastName, err)
	})
}

func TestDeleteContact(t *testing.T) {
	validContact := &definition.Contact{
		ID:        primitive.NewObjectID(),
		FirstName: "valid",
		LastName:  "contact",
		Phone:     "0545454524",
		Address:   "Tel Aviv",
	}
	const expectedDeleted int64 = 1
	const nothingDeleted int64 = 0
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should delete existing contact", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(validContact)
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "acknowledged", Value: true},
			{Key: "n", Value: expectedDeleted}})
		deletedCount, _, err := phoneBookMock.DeleteContact(validContact.ID.String()[10:34])
		assert.Nil(t, err)
		assert.Equal(t, expectedDeleted, deletedCount, "Should delete exactly one contact")
	})

	mt.Run("should not delete wrong ID format", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "acknowledged", Value: false},
			{Key: "n", Value: nothingDeleted}})
		deletedCount, _, err := phoneBookMock.DeleteContact("1234567")
		assert.EqualErrorf(t, err, "the provided hex string is not a valid ObjectID", "wrong ID format")
		assert.Equal(t, -1, int(deletedCount), "got wrong ID format")
	})

	mt.Run("should not delete not existing contact", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "acknowledged", Value: false},
			{Key: "n", Value: nothingDeleted}})
		deletedCount, _, err := phoneBookMock.DeleteContact("123412341234123412341234")
		assert.Nil(t, err)
		assert.Equal(t, deletedCount, nothingDeleted, "Should not delete not existing contact")
	})
}

func TestEditContact(t *testing.T) {
	contact := &definition.Contact{
		ID:        primitive.NewObjectID(),
		FirstName: "bobo",
		LastName:  "dag",
		Phone:     "0545454524",
		Address:   "Tel Aviv",
	}
	const expectedUpdated int64 = 1
	const nothingUpdated int64 = 0

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should edit existing contact", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(contact)
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: expectedUpdated},
			bson.E{Key: "nModified", Value: expectedUpdated},
		))

		updatedCount, _, err := phoneBookMock.UpdateContact(contact.ID.String()[10:34], &definition.Contact{FirstName: "changed"})
		assert.Nil(t, err)
		assert.Equal(t, expectedUpdated, updatedCount, "Should update exactly one contact")
	})

	mt.Run("should not edit wrong format ID", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: nothingUpdated},
			bson.E{Key: "nModified", Value: nothingUpdated},
		))

		deletedCount, _, err := phoneBookMock.UpdateContact("1234567", &definition.Contact{FirstName: "changed"})
		assert.EqualErrorf(t, err, "the provided hex string is not a valid ObjectID", "wrong ID format")
		assert.Equal(t, -1, int(deletedCount), "got wrong ID format")
	})

	mt.Run("should not edit not existing contact", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: nothingUpdated},
			bson.E{Key: "nModified", Value: nothingUpdated},
		))

		updatedCount, _, err := phoneBookMock.UpdateContact("123412341234123412341234", &definition.Contact{FirstName: "changed"})
		assert.Nil(t, err)
		assert.Equal(t, updatedCount, nothingUpdated, "Should not delete not existing contact")
	})

	mt.Run("should not edit without id", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: nothingUpdated},
			bson.E{Key: "nModified", Value: nothingUpdated},
		))

		updatedCount, _, err := phoneBookMock.UpdateContact("", &definition.Contact{FirstName: "changed"})
		assert.EqualErrorf(t, err, ErrorMissingID, "missing ID")
		assert.Equal(t, updatedCount, nothingUpdated, "Should not delete not existing contact")
	})
}

func TestSearchContact(t *testing.T) {
	contacts := []*definition.Contact{
		{
			ID:        primitive.NewObjectID(),
			FirstName: "bobo",
			LastName:  "dag",
			Phone:     "0545454524",
			Address:   "Tel Aviv",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "jojo",
			LastName:  "hey",
			Phone:     "0541112223",
			Address:   "Tel Aviv",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "gogo",
			LastName:  "vivi",
			Phone:     "747455234",
			Address:   "stam address",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "gogo",
			LastName:  "vava",
			Phone:     "0525425452",
			Address:   "stam different address",
		},
	}

	const expectedFound int = 1
	const nothingFound int = 0

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should find one contact by name", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(contacts[1])
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch,
			bson.D{
				{Key: "ID", Value: contacts[1].ID},
				{Key: "FirstName", Value: contacts[1].FirstName},
				{Key: "LastName", Value: contacts[1].LastName},
				{Key: "Phone", Value: contacts[1].Phone},
				{Key: "Address", Value: contacts[1].Address},
			}))
		values := url.Values{
			"firstName": []string{"jojo"},
		}
		foundedContacts, _, err := phoneBookMock.SearchContact(values)
		assert.Nil(t, err)
		assert.Equal(t, expectedFound, len(foundedContacts), "Should find exactly one contact")
	})

	mt.Run("should find one contact by phone", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(contacts[3])
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: contacts[3].ID},
				{Key: "firstName", Value: contacts[3].FirstName},
				{Key: "lastName", Value: contacts[3].LastName},
				{Key: "phone", Value: contacts[3].Phone},
				{Key: "address", Value: contacts[3].Address},
			}))
		values := url.Values{
			"phone": []string{"0525425452"},
		}
		foundedContacts, _, err := phoneBookMock.SearchContact(values)
		assert.Nil(t, err)
		assert.Equal(t, expectedFound, len(foundedContacts), "Should find exactly one contact")
		assert.Equal(t, foundedContacts[0].ID, contacts[3].ID)
		assert.Equal(t, foundedContacts[0].FirstName, contacts[3].FirstName)
		assert.Equal(t, foundedContacts[0].LastName, contacts[3].LastName)
		assert.Equal(t, foundedContacts[0].Phone, contacts[3].Phone)
		assert.Equal(t, foundedContacts[0].Address, contacts[3].Address)
	})

	mt.Run("should find multiple contact", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(contacts[2])
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err = phoneBookMock.AddContact(contacts[3])
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: contacts[2].ID},
				{Key: "firstName", Value: contacts[2].FirstName},
				{Key: "lastName", Value: contacts[2].LastName},
				{Key: "phone", Value: contacts[2].Phone},
				{Key: "address", Value: contacts[2].Address},
			}, bson.D{
				{Key: "_id", Value: contacts[3].ID},
				{Key: "firstName", Value: contacts[3].FirstName},
				{Key: "lastName", Value: contacts[3].LastName},
				{Key: "phone", Value: contacts[3].Phone},
				{Key: "address", Value: contacts[3].Address},
			}))
		values := url.Values{
			"firstName": []string{"gogo"},
		}
		foundedContacts, _, err := phoneBookMock.SearchContact(values)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(foundedContacts), "Should find two contact")
		assert.Equal(t, foundedContacts[0].ID, contacts[2].ID)
		assert.Equal(t, foundedContacts[0].FirstName, contacts[2].FirstName)
		assert.Equal(t, foundedContacts[0].LastName, contacts[2].LastName)
		assert.Equal(t, foundedContacts[0].Phone, contacts[2].Phone)
		assert.Equal(t, foundedContacts[0].Address, contacts[2].Address)
		assert.Equal(t, foundedContacts[1].ID, contacts[3].ID)
		assert.Equal(t, foundedContacts[1].FirstName, contacts[3].FirstName)
		assert.Equal(t, foundedContacts[1].LastName, contacts[3].LastName)
		assert.Equal(t, foundedContacts[1].Phone, contacts[3].Phone)
		assert.Equal(t, foundedContacts[1].Address, contacts[3].Address)
	})

	mt.Run("should find one contact by phone and address", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(contacts[0])
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: contacts[0].ID},
				{Key: "firstName", Value: contacts[0].FirstName},
				{Key: "lastName", Value: contacts[0].LastName},
				{Key: "phone", Value: contacts[0].Phone},
				{Key: "address", Value: contacts[0].Address},
			}))
		values := url.Values{
			"phone":   []string{"0545454524"},
			"address": []string{"Tel Aviv"},
		}
		foundedContacts, _, err := phoneBookMock.SearchContact(values)
		assert.Nil(t, err)
		assert.Equal(t, expectedFound, len(foundedContacts), "Should find exactly one contact")
		assert.Equal(t, foundedContacts[0].ID, contacts[0].ID)
		assert.Equal(t, foundedContacts[0].FirstName, contacts[0].FirstName)
		assert.Equal(t, foundedContacts[0].LastName, contacts[0].LastName)
		assert.Equal(t, foundedContacts[0].Phone, contacts[0].Phone)
		assert.Equal(t, foundedContacts[0].Address, contacts[0].Address)
	})

	mt.Run("should not find contact by address and not existing phone", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(contacts[0])
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch))
		values := url.Values{
			"phone":   []string{"0000000000"},
			"address": []string{"Tel Aviv"},
		}
		foundedContacts, _, err := phoneBookMock.SearchContact(values)
		assert.Nil(t, err)
		assert.Equal(t, nothingFound, len(foundedContacts), "Should not found contact")
	})

	mt.Run("should not found not existing contact", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, _, err := phoneBookMock.AddContact(contacts[0])
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch))
		values := url.Values{
			"firstName": []string{"baba"},
		}
		foundedContacts, _, err := phoneBookMock.SearchContact(values)
		assert.Nil(t, err)
		assert.Equal(t, nothingFound, len(foundedContacts), "Should not found contact")
	})
}

func TestGetContact(t *testing.T) {

	contacts := []*definition.Contact{
		{
			ID:        primitive.NewObjectID(),
			FirstName: "bobo",
			LastName:  "dag",
			Phone:     "0545454524",
			Address:   "Tel Aviv",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "jojo",
			LastName:  "hey",
			Phone:     "0541112223",
			Address:   "Tel Aviv",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "gogo",
			LastName:  "vivi",
			Phone:     "747455234",
			Address:   "stam address",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "gogo",
			LastName:  "vava",
			Phone:     "0525425452",
			Address:   "stam different address",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "hello",
			LastName:  "world",
			Phone:     "0521212121",
			Address:   "some address",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "test1",
			LastName:  "test2",
			Phone:     "0521212122",
			Address:   "Haifa",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "Noa",
			LastName:  "Morag",
			Phone:     "0521212123",
			Address:   "Eilat",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "Hadas",
			LastName:  "Bilu",
			Phone:     "0521212124",
			Address:   "Holon",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "Regev",
			LastName:  "Gor",
			Phone:     "0521212125",
			Address:   "Jerusalem",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "Daniel",
			LastName:  "Cohen",
			Phone:     "0521212126",
			Address:   "Beer Sheva",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "Shir",
			LastName:  "Fori",
			Phone:     "0521212127",
			Address:   "Hod Hasharon",
		},
		{
			ID:        primitive.NewObjectID(),
			FirstName: "abc",
			LastName:  "test23",
			Phone:     "0521212128",
			Address:   "Raanana",
		},
	}
	contactsI := make([]interface{}, len(contacts))
	for i := range contacts {
		contactsI[i] = contacts[i]
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("should return 10 first contacts", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := phoneBookMock.contactsCollection.InsertMany(context.Background(), contactsI)
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: contacts[0].ID},
				{Key: "firstName", Value: contacts[0].FirstName},
				{Key: "lastName", Value: contacts[0].LastName},
				{Key: "phone", Value: contacts[0].Phone},
				{Key: "address", Value: contacts[0].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[1].ID},
				{Key: "firstName", Value: contacts[1].FirstName},
				{Key: "lastName", Value: contacts[1].LastName},
				{Key: "phone", Value: contacts[1].Phone},
				{Key: "address", Value: contacts[1].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[2].ID},
				{Key: "firstName", Value: contacts[2].FirstName},
				{Key: "lastName", Value: contacts[2].LastName},
				{Key: "phone", Value: contacts[2].Phone},
				{Key: "address", Value: contacts[2].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[3].ID},
				{Key: "firstName", Value: contacts[3].FirstName},
				{Key: "lastName", Value: contacts[3].LastName},
				{Key: "phone", Value: contacts[3].Phone},
				{Key: "address", Value: contacts[3].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[4].ID},
				{Key: "firstName", Value: contacts[4].FirstName},
				{Key: "lastName", Value: contacts[4].LastName},
				{Key: "phone", Value: contacts[4].Phone},
				{Key: "address", Value: contacts[4].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[5].ID},
				{Key: "firstName", Value: contacts[5].FirstName},
				{Key: "lastName", Value: contacts[5].LastName},
				{Key: "phone", Value: contacts[5].Phone},
				{Key: "address", Value: contacts[5].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[6].ID},
				{Key: "firstName", Value: contacts[6].FirstName},
				{Key: "lastName", Value: contacts[6].LastName},
				{Key: "phone", Value: contacts[6].Phone},
				{Key: "address", Value: contacts[6].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[7].ID},
				{Key: "firstName", Value: contacts[7].FirstName},
				{Key: "lastName", Value: contacts[7].LastName},
				{Key: "phone", Value: contacts[7].Phone},
				{Key: "address", Value: contacts[7].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[8].ID},
				{Key: "firstName", Value: contacts[8].FirstName},
				{Key: "lastName", Value: contacts[8].LastName},
				{Key: "phone", Value: contacts[8].Phone},
				{Key: "address", Value: contacts[8].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[9].ID},
				{Key: "firstName", Value: contacts[9].FirstName},
				{Key: "lastName", Value: contacts[9].LastName},
				{Key: "phone", Value: contacts[9].Phone},
				{Key: "address", Value: contacts[9].Address},
			},
		))
		result, _, err := phoneBookMock.GetContactWithPagination([]string{"1"})
		assert.Nil(t, err)
		assert.Equal(t, config.Static.LimitPerPage, int64(len(result)), "Should returns 10 contacts")
	})

	mt.Run("should return 2 contacts from the last page", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := phoneBookMock.contactsCollection.InsertMany(context.Background(), contactsI)
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: contacts[10].ID},
				{Key: "firstName", Value: contacts[10].FirstName},
				{Key: "lastName", Value: contacts[10].LastName},
				{Key: "phone", Value: contacts[10].Phone},
				{Key: "address", Value: contacts[10].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[11].ID},
				{Key: "firstName", Value: contacts[11].FirstName},
				{Key: "lastName", Value: contacts[11].LastName},
				{Key: "phone", Value: contacts[11].Phone},
				{Key: "address", Value: contacts[11].Address},
			},
		))
		result, _, err := phoneBookMock.GetContactWithPagination([]string{"2"})
		assert.Nil(t, err)
		assert.Equal(t, 2, len(result), "Should returns 2 contacts")
	})

	mt.Run("should return 10 first contacts when mention an empty page", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := phoneBookMock.contactsCollection.InsertMany(context.Background(), contactsI)
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: contacts[0].ID},
				{Key: "firstName", Value: contacts[0].FirstName},
				{Key: "lastName", Value: contacts[0].LastName},
				{Key: "phone", Value: contacts[0].Phone},
				{Key: "address", Value: contacts[0].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[1].ID},
				{Key: "firstName", Value: contacts[1].FirstName},
				{Key: "lastName", Value: contacts[1].LastName},
				{Key: "phone", Value: contacts[1].Phone},
				{Key: "address", Value: contacts[1].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[2].ID},
				{Key: "firstName", Value: contacts[2].FirstName},
				{Key: "lastName", Value: contacts[2].LastName},
				{Key: "phone", Value: contacts[2].Phone},
				{Key: "address", Value: contacts[2].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[3].ID},
				{Key: "firstName", Value: contacts[3].FirstName},
				{Key: "lastName", Value: contacts[3].LastName},
				{Key: "phone", Value: contacts[3].Phone},
				{Key: "address", Value: contacts[3].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[4].ID},
				{Key: "firstName", Value: contacts[4].FirstName},
				{Key: "lastName", Value: contacts[4].LastName},
				{Key: "phone", Value: contacts[4].Phone},
				{Key: "address", Value: contacts[4].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[5].ID},
				{Key: "firstName", Value: contacts[5].FirstName},
				{Key: "lastName", Value: contacts[5].LastName},
				{Key: "phone", Value: contacts[5].Phone},
				{Key: "address", Value: contacts[5].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[6].ID},
				{Key: "firstName", Value: contacts[6].FirstName},
				{Key: "lastName", Value: contacts[6].LastName},
				{Key: "phone", Value: contacts[6].Phone},
				{Key: "address", Value: contacts[6].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[7].ID},
				{Key: "firstName", Value: contacts[7].FirstName},
				{Key: "lastName", Value: contacts[7].LastName},
				{Key: "phone", Value: contacts[7].Phone},
				{Key: "address", Value: contacts[7].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[8].ID},
				{Key: "firstName", Value: contacts[8].FirstName},
				{Key: "lastName", Value: contacts[8].LastName},
				{Key: "phone", Value: contacts[8].Phone},
				{Key: "address", Value: contacts[8].Address},
			},
			bson.D{
				{Key: "_id", Value: contacts[9].ID},
				{Key: "firstName", Value: contacts[9].FirstName},
				{Key: "lastName", Value: contacts[9].LastName},
				{Key: "phone", Value: contacts[9].Phone},
				{Key: "address", Value: contacts[9].Address},
			},
		))
		result, _, err := phoneBookMock.GetContactWithPagination([]string{""})
		assert.Nil(t, err)
		assert.Equal(t, config.Static.LimitPerPage, int64(len(result)), "Should returns 10 contacts")
	})

	mt.Run("should not return contacts from non existing page", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := phoneBookMock.contactsCollection.InsertMany(context.Background(), contactsI)
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch))
		result, _, err := phoneBookMock.GetContactWithPagination([]string{"4"})
		assert.Nil(t, err)
		assert.Equal(t, 0, len(result), "Should not return contacts")
	})

	mt.Run("should not return contacts from invalid page", func(mt *mtest.T) {
		phoneBookMock := NewMongoPhoneBook(mt.Client)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		_, err := phoneBookMock.contactsCollection.InsertMany(context.Background(), contactsI)
		if err != nil {
			t.Fatalf("Error inserting document: %v", err)
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name()), mtest.FirstBatch))
		result, _, err := phoneBookMock.GetContactWithPagination([]string{"a"})
		assert.NotNil(t, err)
		assert.Equal(t, 0, len(result), "Should not return contacts")
	})
}
