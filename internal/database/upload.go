package database

import (
	"time"

	"gorm.io/gorm"
)

type Upload struct {
	gorm.Model
	OwnerID   uint
	Owner     *Account
	ExpiresAt time.Time
	Filename  string
	MimeType  string
	DomainID  uint
	Domain    Domain
}
