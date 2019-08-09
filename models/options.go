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

type Preloader interface {
	Preloader(db *gorm.DB) *gorm.DB
}

func stringsContains(slice []string, str string) bool {
	for _, val := range slice {
		if val == str {
			return true
		}
	}

	return false
}
