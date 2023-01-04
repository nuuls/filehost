package database

import "gorm.io/gorm"

type DomainAccess struct {
	gorm.Model
	Domain  Domain
	Account Account
}
