package main

import (
	"github.com/nuuls/filehost/internal/api"
	"github.com/nuuls/filehost/internal/database"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	db, err := database.New()
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}

	a := api.New(api.Config{
		DB:   db,
		Log:  log,
		Addr: ":7417",
	})
	err = a.Run()
	if err != nil {
		log.WithError(err).Fatal("Failed to run API")
	}
}
