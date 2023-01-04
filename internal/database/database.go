package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func New() (*Database, error) {
	dsn := "host=localhost user=postgres password=postgrespw dbname=postgres port=49153 sslmode=disable"
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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
