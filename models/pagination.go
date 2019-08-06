package models

import "github.com/jinzhu/gorm"

type Paginator interface {
	Order(db *gorm.DB) *gorm.DB
	Paginate(db *gorm.DB) *gorm.DB
}

type NoPaginator struct{}

func (*NoPaginator) Order(db *gorm.DB) *gorm.DB {
	return db
}

func (*NoPaginator) Paginate(db *gorm.DB) *gorm.DB {
	return db
}
