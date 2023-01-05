package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

type Config struct {
	DSN string
}

func New(cfg *Config) (*Database, error) {
	conn, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
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
