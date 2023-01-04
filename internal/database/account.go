package database

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Username      string
	Password      string
	APIKey        string
	Status        string
	DefaultExpiry *time.Duration
	DefaultDomain *Domain
}
