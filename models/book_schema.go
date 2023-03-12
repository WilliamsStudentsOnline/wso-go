package models

import (
	sql_types "github.com/WilliamsStudentsOnline/wso-go/lib/sql_types"
)

type Book struct {
	BaseSchema

	// Title: Book title.
	Title string `gorm:"not null" json:"title"`
	// Subtitle: Book subtitle.
	Subtitle string `json:"subtitle,omitempty"`
	// Authors: The names of the authors and/or editors for this book.
	Authors sql_types.CSV `gorm:"type:varchar(255)" json:"authors,omitempty"`
	// Publisher: Publisher of this book.
	Publisher string `json:"publisher,omitempty"`
	// ISBN_10: ISBN-10 of this book.
	ISBN_10 string `gorm:"not null" json:"ISBN_10"`
	// ISBN_13: ISBN-13 of this book.
	ISBN_13 string `gorm:"not null" json:"ISBN_13"`
	// InfoLink: URL to view information about this book on the Google
	// Books site.
	InfoLink string `gorm:"size:65535" json:"infoLink,omitempty"`
	// ImageLink: Image link for book cover (Medium)
	ImageLink string `gorm:"size:65535" json:"imageLinks,omitempty"`

	// Has many Book Listings
	BookListings []*BookListing `gorm:"foreignkey:BookID" json:"bookListings,omitempty"`
	// Many2Many courses
	Courses []*Course `gorm:"many2many:course_book" json:"courses,omitempty"`
}

func (*Book) TableName() string {
	return "books"
}
