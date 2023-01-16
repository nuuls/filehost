package database

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

type Config struct {
	DSN      string
	Log      logrus.FieldLogger
	LogLevel logger.LogLevel
}

func New(cfg *Config) (*Database, error) {
	conn, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(cfg.Log, logger.Config{
			LogLevel:                  cfg.LogLevel,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return nil, err
	}

	db := &Database{
		db: conn,
	}

	err = db.db.AutoMigrate(
		&Account{},
		&Domain{},
		&DomainAccess{},
		&Upload{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
