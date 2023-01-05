package database

import (
	"time"

	"gorm.io/gorm"
)

type Upload struct {
	gorm.Model
	OwnerID   *uint
	Owner     *Account
	ExpiresAt time.Time
	Filename  string
	MimeType  string
	DomainID  uint
	Domain    Domain
}

func (db *Database) CreateUpload(upload Upload) (*Upload, error) {
	res := db.db.Create(&upload)
	if res.Error != nil {
		return nil, res.Error
	}
	return &upload, nil
}

func (db *Database) GetUploadsByAccount(accountID uint, limit, offset int) ([]*Upload, error) {
	out := []*Upload{}
	res := db.db.Limit(limit).Offset(offset).Find(&out, "owner_id = ?", accountID)
	if res.Error != nil {
		return nil, res.Error
	}
	return out, nil
}
