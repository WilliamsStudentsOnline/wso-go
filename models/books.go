package models

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/isbn"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
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
	Isbn      *string `json:"isbn" form:"isbn" binding:"omitempty,isbn"`
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
	if o.Isbn != nil {
		db = m.withIsbn(isbn.ConvertIsbn10to13(isbn.CleanISBN(*o.Isbn)))(db)
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

func (m *BookModel) withIsbn(isbn string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("isbn = ?", isbn)
	}
}

type BookIdentifier struct {
	Id   *uint
	Isbn *string
}

func (m *BookModel) DoesBookExist(identifier BookIdentifier) (exists bool, err error) {
	var count int
	if identifier.Id != nil {
		err = m.DB.Model(&Book{}).Where("books.id = ?", *identifier.Id).Count(&count).Error
	} else if identifier.Isbn != nil {
		err = m.DB.Model(&Book{}).Where("books.isbn = ?", isbn.ConvertIsbn10to13(isbn.CleanISBN(*identifier.Isbn))).Count(&count).Error
	} else {
		err = fmt.Errorf("missing valid identifier")
	}
	exists = count > 0
	return
}

func (m *BookModel) CreateBook(b *Book) (err error) {
	err = m.DB.FirstOrCreate(b, &Book{Isbn: b.Isbn}).Error
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
		err = tx.First(course, courseID).Error
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
