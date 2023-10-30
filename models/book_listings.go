package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/isbn"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type BookListingModel struct {
	*BaseModel
}

func NewBookListingModel(db *gorm.DB, log *zap.SugaredLogger) *BookListingModel {
	return &BookListingModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllBookListingsOptions struct {
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// Filters
	BookID       *uint        `json:"bookID" form:"bookID"`
	CourseID     *uint        `json:"courseID" form:"courseID"`
	UserID       *uint        `json:"userID" form:"userID"`
	Isbn         *string      `json:"isbn" form:"isbn" binding:"omitempty,isbn"`
	MinCondition *Condition   `json:"minCondition" form:"minCondition"`
	MaxCondition *Condition   `json:"maxCondition" form:"maxCondition"`
	ListingType  *ListingType `json:"listingType" form:"listingType"`
}

func (o *GetAllBookListingsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("book_listings.created_at DESC", true)
}

func (o *GetAllBookListingsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}

	return db
}

func (m *BookListingModel) GetAllBookListings(c *[]*BookListing, opts *GetAllBookListingsOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}

	// Do db query
	err = db.Preload("User").Find(c).Error
	return
}

func (o *GetAllBookListingsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Paginate(db)

	m := NewBookListingModel(db.New(), nil)

	// If we filter by book ID, we ignore the isbn
	if o.BookID != nil {
		db = m.withBookID(*o.BookID)(db)
	} else {
		if o.Isbn != nil {
			db = m.withIsbn(isbn.CleanISBN(*o.Isbn))(db)
		}
	}
	if o.CourseID != nil {
		db = m.withCourse(*o.CourseID)(db)
	}
	if o.ListingType != nil {
		db = m.withListingType(*o.ListingType)(db)
	}
	if o.UserID != nil {
		db = m.withUser(*o.UserID)(db)
	}
	if o.MinCondition != nil {
		db = m.withMinCondition(*o.MinCondition)(db)
	}
	if o.MaxCondition != nil {
		db = m.withMaxCondition(*o.MaxCondition)(db)
	}

	return db
}

func (m *BookListingModel) withBookID(bookID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("book_id = ?", bookID)
	}
}

func (m *BookListingModel) withCourse(courseID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"book_listings.book_id in (?)",
			m.DB.Table("course_book").Select("book_id").Where(
				"course_id = ?", courseID,
			).QueryExpr(),
		)
	}
}

func (m *BookListingModel) withUser(userID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

func (m *BookListingModel) withIsbn(isbn string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"book_listings.book_id in (?)",
			m.DB.Table("books").Select("book_id").Where(
				"isbn = ?", isbn,
			).QueryExpr(),
		)
	}
}

func (m *BookListingModel) withListingType(listingType ListingType) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("listing_type = ?", listingType)
	}
}

func (m *BookListingModel) withMinCondition(minCondition Condition) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("condition >= ?", minCondition)
	}
}

func (m *BookListingModel) withMaxCondition(maxCondition Condition) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("condition <= ?", maxCondition)
	}
}

func (m *BookListingModel) CreateBookListing(b *BookListing) (err error) {
	err = m.DB.Create(b).Error
	if err != nil {
		return err
	}
	err = m.DB.Find(b).Error
	return
}

func (m *BookListingModel) UpdateBookListing(b *BookListing) (err error) {
	err = m.DB.Save(b).Error
	return
}

func (m *BookListingModel) GetBookListingByID(id uint, b *BookListing) (err error) {
	err = m.DB.First(b, id).Error
	return
}

func (m *BookListingModel) DeleteBookListing(b *BookListing) (err error) {
	err = m.DB.Delete(b).Error
	return
}
