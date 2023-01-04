package database

import "gorm.io/gorm"

type DomainAccess struct {
	gorm.Model
	DomainID  uint
	Domain    Domain
	AccountID uint
	Account   Account
}
