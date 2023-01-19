package database

import (
	"time"

	"gorm.io/gorm"
)

type Upload struct {
	gorm.Model
	OwnerID      *uint `gorm:"index"`
	Owner        *Account
	UploaderIP   string
	Filename     string `gorm:"uniqueIndex"`
	MimeType     string
	SizeBytes    uint
	DomainID     uint
	Domain       Domain
	TTLSeconds   *uint
	LastViewedAt time.Time `gorm:"default:now()"`
	Views        uint      `gorm:"default:0; not null"`
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
	res := db.db.
		Joins("Domain").
		Order("id DESC").Limit(limit).Offset(offset).
		Find(&out, "uploads.owner_id = ?", accountID)
	if res.Error != nil {
		return nil, res.Error
	}
	return out, nil
}

func (db *Database) IncUploadViews(filename string) error {
	res := db.db.Model(Upload{}).
		Where("filename = ?", filename).
		Updates(map[string]interface{}{
			"views":          gorm.Expr("views + 1"),
			"last_viewed_at": gorm.Expr("now()"),
		})
	return res.Error
}

func (db *Database) GetUploadByFilename(filename string) (*Upload, error) {
	upload := &Upload{}
	res := db.db.First(upload, "filename = ?", filename)
	if res.Error != nil {
		return nil, res.Error
	}
	return upload, nil
}

func (db *Database) DeleteUpload(id uint) error {
	res := db.db.Delete(&Upload{}, "id = ?", id)
	return res.Error
}
