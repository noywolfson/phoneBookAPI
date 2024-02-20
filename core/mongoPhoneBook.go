package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
	"phoneBook/config"
	"phoneBook/definition"
	"regexp"
	"strconv"
)

var (
	onlyDigitsRegex       = regexp.MustCompile(`^[0-9]+$`)
	onlyLettersRegex      = regexp.MustCompile(`^[a-zA-Z]+$`)
	ErrorMissingFirstName = "can't add contact without first name"
	ErrorMissingPhone     = "can't add contact without phone number"
	ErrorMissingID        = "doesn't sent contact id"
	ErrorInvalidPhone     = "invalid phone number. phone should include digits only"
	ErrorInvalidFirstName = "invalid first name. name should include letters only"
	ErrorInvalidLastName  = "invalid last name. name should include letters only"
	BadRequest            = "BadRequest"
	InternalServerError   = "InternalServerError"
)

type MongoPhoneBook struct {
	client             *mongo.Client
	contactsCollection *mongo.Collection
	limitPerPage       int64
}

func NewMongoPhoneBook(mongoClient *mongo.Client) *MongoPhoneBook {
	contactsCollection := mongoClient.Database(config.Static.MongoDBName).Collection(config.Static.MongoCollectionName)
	return &MongoPhoneBook{
		client:             mongoClient,
		contactsCollection: contactsCollection,
		limitPerPage:       config.Static.LimitPerPage,
	}
}

func (pb *MongoPhoneBook) GetContactWithPagination(pageParam []string) ([]*definition.Contact, string, error) {
	page, err := validatePageParam(pageParam)
	if err != nil {
		return nil, BadRequest, err
	}
	findOptions := *options.Find()
	findOptions.SetLimit(config.Static.LimitPerPage)
	findOptions.SetSkip(int64(page-1) * config.Static.LimitPerPage)
	cursor, err := pb.contactsCollection.Find(context.TODO(), bson.M{}, &findOptions)
	if err != nil {
		return nil, InternalServerError, err
	}
	defer cursor.Close(context.TODO())
	var contacts []*definition.Contact
	for cursor.Next(context.Background()) {
		var contact *definition.Contact
		err := cursor.Decode(&contact)
		if err != nil {
			return nil, BadRequest, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, "", nil
}

func validatePageParam(pageParam []string) (int, error) {
	if len(pageParam) == 0 {
		return 1, nil
	}
	pageStr := pageParam[0]
	if pageStr == "" {
		return 1, nil
	}
	if pageStr == "0" {
		return -1, errors.New("page number must be positive")
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return -1, err
	}
	return page, nil
}

func (pb *MongoPhoneBook) SearchContact(query url.Values) ([]*definition.Contact, string, error) {
	filter := bson.M{}
	for key, value := range query {
		filter[key] = value[0]
	}
	if len(query) == 0 {
		return pb.GetContactWithPagination([]string{"1"})
	}
	cursor, err := pb.contactsCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, InternalServerError, err
	}
	defer cursor.Close(context.TODO())
	var contacts []*definition.Contact
	for cursor.Next(context.Background()) {
		var contact *definition.Contact
		err := cursor.Decode(&contact)
		if err != nil {
			return nil, BadRequest, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, "", nil
}

func (pb *MongoPhoneBook) DeleteContact(idParam string) (int64, string, error) {
	if idParam == "" {
		logrus.Println("doesn't sent contact id to delete")
		return 0, BadRequest, errors.New(ErrorMissingID)
	}
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return -1, BadRequest, err
	}
	filter := bson.M{"_id": id}
	deleteResult, err := pb.contactsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return -1, InternalServerError, err
	}
	if deleteResult.DeletedCount == 0 {
		return 0, "", nil
	}
	return deleteResult.DeletedCount, "", nil
}

func (pb *MongoPhoneBook) UpdateContact(idParam string, contact *definition.Contact) (int64, string, error) {
	if idParam == "" {
		logrus.Println("doesn't sent contact id to delete")
		return 0, BadRequest, errors.New(ErrorMissingID)
	}
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return -1, BadRequest, err
	}
	filter := bson.M{"_id": id}
	updatedCount, err := pb.contactsCollection.UpdateOne(context.Background(), filter, bson.M{"$set": contact})
	if err != nil {
		return -1, InternalServerError, err
	}
	if updatedCount.ModifiedCount == 0 {
		return 0, "", nil
	}
	return updatedCount.ModifiedCount, "", nil
}

func (pb *MongoPhoneBook) AddContact(contact *definition.Contact) (string, string, error) {
	err := validateContact(contact)
	if err != nil {
		return "", BadRequest, err
	}
	result, err := pb.contactsCollection.InsertOne(context.Background(), contact)
	if err != nil {
		return "", InternalServerError, err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", InternalServerError, err
	}
	return fmt.Sprintf("Inserted ID: %s", id.String()[10:34]), "", nil
}

func validateContact(contact *definition.Contact) error {
	if contact.FirstName == "" {
		return errors.New(ErrorMissingFirstName)
	}
	if !onlyLettersRegex.MatchString(contact.FirstName) {
		return errors.New(ErrorInvalidFirstName)
	}
	if contact.LastName != "" && !onlyLettersRegex.MatchString(contact.LastName) {
		return errors.New(ErrorInvalidLastName)
	}
	if contact.Phone == "" {
		return errors.New(ErrorMissingPhone)
	}
	if !onlyDigitsRegex.MatchString(contact.Phone) {
		return errors.New(ErrorInvalidPhone)
	}
	return nil
}
