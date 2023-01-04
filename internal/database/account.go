package database

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Username        string
	Password        string
	APIKey          string
	Status          string
	DefaultExpiry   *time.Duration
	DefaultDomainID uint
	DefaultDomain   *Domain
}

func (db *Database) CreateAccount(acc Account) (*Account, error) {
	res := db.db.Create(&acc)
	if res.Error != nil {
		return nil, res.Error
	}
	return &acc, nil
}
