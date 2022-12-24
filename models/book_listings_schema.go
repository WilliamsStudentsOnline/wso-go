package models

type BookListing struct {
	BaseSchema

	// Book Info
	// Title: Book title.
	Title *string `gorm:"not null" json:"title,omitempty"`
	// Subtitle: Book subtitle.
	Subtitle *string `json:"subtitle,omitempty"`
	// Authors: The names of the authors and/or editors for this book.
	Authors *string `json:"authors,omitempty"`
	// Publisher: Publisher of this book.
	Publisher *string `json:"publisher,omitempty"`
	// ISBN_10: ISBN-10 of this book.
	ISBN_10 *string `gorm:"not null" json:"ISBN_10"`
	// ISBN_13: ISBN-13 of this book.
	ISBN_13 *string `gorm:"not null" json:"ISBN_13"`
	// InfoLink: URL to view information about this book on the Google
	// Books site.
	InfoLink *string `gorm:"size:65535" json:"infoLink,omitempty"`
	// ImageLink: Image link for book cover
	ImageLink *string `gorm:"size:65535" json:"imageLinks,omitempty"`

	// Listing Info

	// Belongs to user (student)
	UserID uint  `gorm:"index:index_book_listings_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	// Belongs to course
	CourseID uint    `gorm:"index:index_book_listings_on_course_id;not null" json:"courseID"`
	Course   *Course `json:"course,omitempty"`

	Condition   *string `json:"condition"`
	Description *string `gorm:"size:65535" json:"description"`

	// True -> Offering to buy, False -> Offering to sell
	Buy bool `json:"-"`
}

func (*BookListing) TableName() string {
	return "book_listings"
}
