package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"phoneBook/config"
	"phoneBook/definition"
)

type httpHandlerStruct struct {
	phoneBook *definition.IPhoneBook
}

var httpHandler httpHandlerStruct

func initHttpHandler(phoneBook *definition.IPhoneBook) {
	if phoneBook == nil {
		logrus.Fatal("can not init httpHandler - phoneBook is nil")
	}
	httpHandler = httpHandlerStruct{
		phoneBook: phoneBook,
	}
}

// @Summary Get contacts with pagination
// @Description Retrieve contacts with pagination support, up to 10 contacts for each page
// @Produce json
// @Param page query string false "Page number (default 1)"
// @Success 200 {array} definition.Contact
// @Router /contact [get]
func (h *httpHandlerStruct) GetContactWithPagination(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageParam := query["page"]
	result, err := (*h.phoneBook).GetContactWithPagination(pageParam)
	if err != nil {
		h.handleError(err, w, http.StatusInternalServerError)
		return
	}
	response, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// @Summary Add a new contact
// @Description Add a new contact to the phone book
// @Accept json
// @Produce json
// @Param contact body definition.Contact true "Contact object that needs to be added"
// @Success 200 {string} string "Contact added successfully"
// @Router /contact [post]
func (h *httpHandlerStruct) AddContact(w http.ResponseWriter, r *http.Request) {
	contact, err := h.decodeContact(r.Body)
	if err != nil {
		h.handleError(err, w, http.StatusInternalServerError)
		return
	}
	result, err := (*h.phoneBook).AddContact(contact)
	if err != nil {
		h.handleError(err, w, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

// @Summary Delete a contact by ID
// @Description Deletes a contact by its ID
// @Param id path string true "Contact ID (24 characters)"
// @Success 200 {string} string "Message indicating successful deletion"
// @Failure 400 {string} string "Bad request: invalid ID format"
// @Router /contact/delete/{id} [delete]
func (h *httpHandlerStruct) DeleteContact(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deleteCount, err := (*h.phoneBook).DeleteContact(params["id"])
	if err != nil {
		h.handleError(err, w, http.StatusInternalServerError)
		return
	}
	var response []byte
	if deleteCount == 0 {
		response, _ = json.Marshal("not found document to delete")
	} else {
		response, _ = json.Marshal(fmt.Sprintf("deleted %d document successfully", deleteCount))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// @Summary Update a contact by ID
// @Description Updates a contact by its ID
// @Param id path string true "Contact ID (24 characters)"
// @Param contact body definition.Contact true "Contact details to update"
// @Success 200 {string} string "Message indicating successful update"
// @Failure 400 {string} string "Bad request: invalid ID format"
// @Failure 404 {string} string "Not found: no document found to update"
// @Router /contact/edit/{id} [put]
func (h *httpHandlerStruct) UpdateContact(w http.ResponseWriter, r *http.Request) {
	updatedContact, err := h.decodeContact(r.Body)
	if err != nil {
		h.handleError(err, w, http.StatusInternalServerError)
		return
	}
	params := mux.Vars(r)
	updatedCount, err := (*h.phoneBook).UpdateContact(params["id"], updatedContact)
	if err != nil {
		h.handleError(err, w, http.StatusInternalServerError)
		return
	}
	var response []byte
	if updatedCount == 0 {
		response, _ = json.Marshal("not found document to edit")
	} else {
		response, _ = json.Marshal(fmt.Sprintf("edited %d document successfully", updatedCount))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// @Summary Search contacts
// @Description Searches for contacts based on parameters (firstName, lastName, phone, address). If no parameters are provided, returns all contacts.
// @Param firstName query string false "firsName"
// @Param lastName query string false "lastName"
// @Param phone query string false "phone"
// @Param address query string false "address"
// @Success 200 {array} definition.Contact
// @Router /contact/search [get]
func (h *httpHandlerStruct) SearchContact(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	contacts, err := (*h.phoneBook).SearchContact(query)
	if err != nil {
		h.handleError(err, w, http.StatusInternalServerError)
		return
	}
	response, _ := json.Marshal(contacts)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (h *httpHandlerStruct) handleError(err error, w http.ResponseWriter, status int) {
	logrus.WithError(err).Error()
	w.WriteHeader(status)
	response, _ := json.Marshal(err.Error())
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	return
}

func (h *httpHandlerStruct) decodeContact(r io.ReadCloser) (*definition.Contact, error) {
	var contact *definition.Contact
	err := json.NewDecoder(r).Decode(&contact)
	if err != nil {
		return nil, err
	}
	isValidInput := h.validateContactSizeInput(contact)
	if !isValidInput {
		return nil, errors.New("too big contact field")
	}
	return contact, nil
}

func (h *httpHandlerStruct) validateContactSizeInput(contact *definition.Contact) bool {
	if contact == nil {
		return false
	}
	return len(contact.FirstName) <= config.Static.MaxSizeProperty &&
		len(contact.LastName) <= config.Static.MaxSizeProperty &&
		len(contact.Phone) <= config.Static.MaxSizeProperty &&
		len(contact.Address) <= config.Static.MaxSizeProperty
}