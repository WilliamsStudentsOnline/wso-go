package models

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	ISBN_INVALID = -1
	ISBN_10      = 10
	ISBN_13      = 13
)

type BookModel struct {
	*BaseModel
}

func NewBookModel(db *gorm.DB, log *zap.SugaredLogger) *BookModel {
	return &BookModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllBooksOptions struct {
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	Title     *string `json:"title" form:"title"`
	Publisher *string `json:"publisher,omitempty" form:"publisher"`
	ISBN_10   *string `json:"isbn10" form:"isbn10" binding:"omitempty,isbn10"`
	ISBN_13   *string `json:"isbn13" form:"isbn13" binding:"omitempty,isbn13"`
}

func (m *BookModel) GetBookByID(id uint, b *Book) (err error) {
	err = m.DB.Where(id).First(b).Error
	return
}

func (o *GetAllBooksOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("books.created_at DESC", true)
}

func (o *GetAllBooksOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

func (m *BookModel) GetAllBooks(c *[]*Book, opts *GetAllBooksOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}

	err = db.Find(c).Error
	return
}

func (o *GetAllBooksOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Paginate(db)

	m := NewBookModel(db.New(), nil)

	if o.Title != nil {
		db = m.withTitle(*o.Title)(db)
	}
	if o.Publisher != nil {
		db = m.withPublisher(*o.Publisher)(db)
	}
	if o.ISBN_10 != nil {
		db = m.withISBN_10(CleanISBN(*o.ISBN_10))(db)
	}
	if o.ISBN_13 != nil {
		db = m.withISBN_13(CleanISBN(*o.ISBN_13))(db)
	}
	return db
}

func (m *BookModel) withTitle(title string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("title LIKE ?", title+"%")
	}
}

func (m *BookModel) withPublisher(publisher string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("publisher = ?", publisher)
	}
}

func (m *BookModel) withISBN_10(ISBN_10 string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("isbn10 = ?", ISBN_10)
	}
}

func (m *BookModel) withISBN_13(ISBN_13 string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("isbn13 = ?", ISBN_13)
	}
}

func (m *BookModel) DoesBookExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Book{}).Where("books.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

func (m *BookModel) DoesBookExistByISBN10(ISBN_10 uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Book{}).Where("books.isbn10 = ?", ISBN_10).Count(&count).Error
	exists = count > 0
	return
}

func (m *BookModel) DoesBookExistByISBN13(ISBN_13 uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Book{}).Where("books.isbn13 = ?", ISBN_13).Count(&count).Error
	exists = count > 0
	return
}

func (m *BookModel) CreateBook(b *Book) (err error) {
	err = m.DB.FirstOrCreate(b, &Book{ISBN_10: b.ISBN_10, ISBN_13: b.ISBN_13}).Error
	if err != nil {
		return err
	}

	err = m.DB.Find(b).First(b).Error
	return
}

func (m *BookModel) AddCoursesToBook(id uint, courseIDs *[]uint) (b *Book, err error) {
	book := new(Book)
	err = m.DB.First(&book, id).Error
	if err != nil {
		return nil, err
	}

	tx := m.DB.Begin()
	if err = tx.Error; err != nil {
		return
	}

	for _, courseID := range *courseIDs {
		course := new(Course)
		err = m.DB.First(course, courseID).Error
		if err != nil {
			tx.Rollback()
			if gorm.IsRecordNotFoundError(err) {
				return nil, lib.ErrorBookCourseNotFound
			}
			return nil, err
		}

		err = tx.Model(&book).Association("Courses").Append(course).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return
	}
	return book, nil
}

func (m *BookModel) DeleteBook(b *Book) (err error) {
	err = m.DB.Delete(b).Error
	return
}

func CleanISBN(isbn string) string {
	return strings.ReplaceAll(strings.ReplaceAll(isbn, " ", ""), "-", "")
}
