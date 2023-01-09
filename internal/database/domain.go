package database

import "gorm.io/gorm"

type DomainStatus string

const (
	DomainStatusPending DomainStatus = "pending"
)

type Domain struct {
	gorm.Model
	OwnerID          uint
	Owner            *Account
	Domain           string
	AccessRequired   bool
	AllowedMimeTypes []string `gorm:"serializer:json"`

	Status DomainStatus
}

func (db *Database) CreateDomain(domain Domain) (*Domain, error) {
	res := db.db.Create(&domain)
	if res.Error != nil {
		return nil, res.Error
	}
	return &domain, nil
}

func (db *Database) GetDomains(limit, offset int) ([]*Domain, error) {
	out := []*Domain{}
	res := db.db.Limit(limit).Offset(offset).Find(&out)
	if res.Error != nil {
		return nil, res.Error
	}
	return out, nil
}

func (db *Database) GetDomainByID(id uint) (*Domain, error) {
	out := &Domain{}
	res := db.db.First(&out, "id = ?", id)
	if res.Error != nil {
		return nil, res.Error
	}
	return out, nil
}
