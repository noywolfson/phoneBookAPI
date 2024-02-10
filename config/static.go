package config

import (
	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
)

var Static = struct {
	HTTPServerPort      string `env:"HTTP_SERVER_PORT" envDefault:":8080"`
	LimitPerPage        int64  `end:"LIMIT_PER_PAGE" envDefault:"10"`
	MongoURI            string `end:"MONGO_URI" envDefault:"mongodb://mongo:27017"`
	MongoDBName         string `end:"MONGO_DB" envDefault:"phoneBook"`
	MongoCollectionName string `end:"MONGO_COLLECTION" envDefault:"contacts"`
	MaxSizeProperty     int    `end:"MAX_SIZE_PROPERTY" envDefault:"100"`
}{}

func init() {
	err := env.Parse(&Static)
	if err != nil {
		logrus.WithError(err).Fatal("Could not load static configuration")
	}
}
