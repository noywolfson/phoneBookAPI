package server

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"phoneBook/config"
	"phoneBook/definition"
)

var httpServer *http.Server

func StartHTTP(phoneBook *definition.IPhoneBook) *http.Server {
	router := mux.NewRouter()
	initHttpHandler(phoneBook)
	registerRoutes(router)
	httpServer = &http.Server{
		Addr:    config.Static.HTTPServerPort,
		Handler: router,
	}
	go listenAndServe(httpServer)
	return httpServer
}

func listenAndServe(server *http.Server) {
	logrus.Infof("Starting http server on addr %v", config.Static.HTTPServerPort)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logrus.WithError(err).Fatal("failed to start http server")
	}
}

// @title Phonebook API
// @description Phonebook API allows users to manage contacts, including add, delete, edit, get with pagination and search
func registerRoutes(router *mux.Router) {
	router.HandleFunc("/contact", httpHandler.GetContactWithPagination).Methods("GET")
	router.HandleFunc("/contact", httpHandler.AddContact).Methods("POST")
	router.HandleFunc("/contact/edit/{id}", httpHandler.UpdateContact).Methods("PUT")
	router.HandleFunc("/contact/delete/{id}", httpHandler.DeleteContact).Methods("DELETE")
	router.HandleFunc("/contact/search", httpHandler.SearchContact).Methods("GET")
	router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
	router.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})
}

func Shutdown() {
	if httpServer != nil {
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
	}
}
