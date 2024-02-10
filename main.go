package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"os/signal"
	"phoneBook/config"
	"phoneBook/core"
	"phoneBook/definition"
	"phoneBook/server"
	"syscall"
	"time"
)

var client *mongo.Client

func main() {
	mongoClient := initDB()
	phoneBook := initPhoneBook(mongoClient)

	server.StartHTTP(&phoneBook)

	// Handle OS signals for graceful shutdown
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	_ = <-gracefulShutdown
	shutDown()
}

func shutDown() {
	log.Println("Shutting down server...")
	server.Shutdown()

	log.Println("Disconnecting MongoDB client...")
	disconnectDB()

	log.Println("Server gracefully stopped")
}

func initDB() *mongo.Client {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(config.Static.MongoURI))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func disconnectDB() {
	if client != nil {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println("Failed to disconnect from MongoDB:", err)
		}
	}
}

func initPhoneBook(mongoClient *mongo.Client) definition.IPhoneBook {
	return core.NewMongoPhoneBook(mongoClient)
}
