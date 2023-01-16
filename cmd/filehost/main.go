package main

import (
	"github.com/nuuls/filehost/internal/api"
	"github.com/nuuls/filehost/internal/config"
	"github.com/nuuls/filehost/internal/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logrus.New()
	log.Level = cfg.LogLevel

	dbLogLevel := logger.Info
	if cfg.LogLevel != logrus.DebugLevel {
		dbLogLevel = logger.Error
	}

	db, err := database.New(&database.Config{
		DSN:      cfg.PostgresDSN,
		Log:      log,
		LogLevel: dbLogLevel,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}

	a := api.New(api.Config{
		DB:     db,
		Log:    log,
		Config: cfg,
	})
	err = a.Run()
	if err != nil {
		log.WithError(err).Fatal("Failed to run API")
	}
}
