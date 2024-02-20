package config

import (
	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
)

var Static = struct {
	HTTPServerPort      string `env:"HTTP_SERVER_PORT" envDefault:":8080"`
	LimitPerPage        int64  `env:"LIMIT_PER_PAGE" envDefault:"10"`
	MongoURI            string `env:"MONGO_URI" envDefault:"mongodb://mongo:27017"`
	MongoDBName         string `env:"MONGO_DB" envDefault:"phoneBook"`
	MongoCollectionName string `env:"MONGO_COLLECTION" envDefault:"contacts"`
	MaxSizeProperty     int    `env:"MAX_SIZE_PROPERTY" envDefault:"100"`
}{}

func init() {
	err := env.Parse(&Static)
	if err != nil {
		logrus.WithError(err).Fatal("Could not load static configuration")
	}
}
