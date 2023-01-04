package database

import "gorm.io/gorm"

type Domain struct {
	gorm.Model
	OwnerID          uint
	Owner            *Account
	Domain           string
	AccessRequired   bool
	AllowedMimeTypes []string `gorm:"serializer:json"`
}
