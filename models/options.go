package models

import "github.com/jinzhu/gorm"

type Options interface {
	Run(db *gorm.DB) *gorm.DB
}

type Preloader interface {
	Preloader(db *gorm.DB) *gorm.DB
}

type Paginator interface {
	Order(db *gorm.DB) *gorm.DB
	Paginate(db *gorm.DB) *gorm.DB
}

type Filterer interface {
	Filter(db *gorm.DB) *gorm.DB
}

type FullOptions interface {
	Options
	Preloader
	Paginator
	Filterer
}

type NoPaginator struct{}

func (*NoPaginator) Order(db *gorm.DB) *gorm.DB {
	return db
}

func (*NoPaginator) Paginate(db *gorm.DB) *gorm.DB {
	return db
}

func stringsContains(slice []string, str string) bool {
	for _, val := range slice {
		if val == str {
			return true
		}
	}

	return false
}
