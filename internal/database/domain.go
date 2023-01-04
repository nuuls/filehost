package database

import "gorm.io/gorm"

type Domain struct {
	gorm.Model
	Owner            Account
	Domain           string
	AccessRequired   bool
	AllowedMimeTypes []string
}
