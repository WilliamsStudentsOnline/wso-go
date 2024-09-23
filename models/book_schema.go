package models

import (
	sql_types "github.com/WilliamsStudentsOnline/wso-go/lib/sql_types"
)

type Book struct {
	BaseSchema

	// Book title.
	Title string `gorm:"not null" json:"title"`
	// Book subtitle.
	Subtitle *string `json:"subtitle,omitempty"`
	// The names of the authors and/or editors for this book.
	Authors sql_types.CSV `gorm:"type:varchar(255)" json:"authors,omitempty" swaggertype:"array,string"`
	// Publisher of this book.
	Publisher *string `json:"publisher,omitempty"`
	// ISBN-13 of this book (all books have this)
	Isbn string `gorm:"not null" json:"isbn13"`
	// URL to view information about this book on the Google Books site.
	InfoLink *string `gorm:"size:65535" json:"infoLink,omitempty"`
	// Image link for book cover (Medium)
	ImageLink *string `gorm:"size:65535" json:"imageLink,omitempty"`

	// Has many Book Listings
	BookListings []*BookListing `gorm:"foreignKey:BookID" json:"bookListings,omitempty"`
	// Many2Many courses
	Courses []*Course `gorm:"many2many:course_book;" json:"courses,omitempty"`
}

func (*Book) TableName() string {
	return "books"
}
