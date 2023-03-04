package main

import (
	"os"

	"github.com/nuuls/filehost/internal/api"
	"github.com/nuuls/filehost/internal/config"
	"github.com/nuuls/filehost/internal/database"
	"github.com/nuuls/filehost/internal/filestore"
	"github.com/nuuls/filehost/internal/filestore/diskstore"
	"github.com/nuuls/filehost/internal/filestore/multistore"
	"github.com/nuuls/filehost/internal/filestore/s3store"
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

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "import":
			err := ImportFilesFromFS(cfg, log, db)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
	}

	a := api.New(api.Config{
		DB: db,
		// Filestore: diskstore.New(cfg.FallbackFilePath),
		// Filestore: s3store.New(cfg),
		Filestore: multistore.New([]filestore.Filestore{
			s3store.New(cfg),
			diskstore.New(cfg.FallbackFilePath),
		}),
		Log:    log,
		Config: cfg,
	})
	err = a.Run()
	if err != nil {
		log.WithError(err).Fatal("Failed to run API")
	}
}
