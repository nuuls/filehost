package database

import (
	"time"

	"gorm.io/gorm"
)

type Upload struct {
	gorm.Model
	Owner     *Account
	ExpiresAt time.Time
	Filename  string
	MimeType  string
	Domain    Domain
}
