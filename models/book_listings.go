package models

import (
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
	CourseID     *uint   `json:"courseID" form:"courseID"`
	UserID       *uint   `json:"userID" form:"userID"`
	ISBN_10      *string `json:"ISBN_10" form:"ISBN_10"`
	ISBN_13      *string `json:"ISBN_13" form:"ISBN_13"`
	Condition    *string `json:"condition" form:"condition"`
	IsBuyListing *bool   `json:"isBuy" form:"isBuy"`
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
	err = db.Find(c).Error
	return
}

func (o *GetAllBookListingsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Paginate(db)

	m := NewBookListingModel(db.New(), nil)

	if o.CourseID != nil {
		db = m.withCourse(*o.CourseID)(db)
	}
	if o.IsBuyListing != nil {
		db = m.withListingType(*o.IsBuyListing)(db)
	}
	if o.ISBN_13 != nil {
		db = m.withISBN13(*o.ISBN_13)(db)
	} else if o.ISBN_10 != nil {
		db = m.withISBN10(*o.ISBN_10)(db)
	}
	if o.UserID != nil {
		db = m.withUser(*o.UserID)(db)
	}

	return db
}

func (m *BookListingModel) withCourse(courseID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"book_listings.id in (?)",
			m.DB.Table("course_bookListing").Select("book_listing_id").Where(
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

func (m *BookListingModel) withISBN10(ISBN_10 string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("isbn_10 = ?", ISBN_10)
	}
}

func (m *BookListingModel) withISBN13(ISBN_13 string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("isbn_13 = ?", ISBN_13)
	}
}

func (m *BookListingModel) withListingType(isBuyListing bool) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_buy_listing = ?", isBuyListing)
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

func (m *BookListingModel) DeleteBookListing(b *BookListing) (err error) {
	err = m.DB.Delete(b).Error
	return
}
