package models

import (
	"strconv"

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

func (m *BookModel) DoesBookExistByID(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Book{}).Where("books.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

func (m *BookModel) DoesBookExistByISBN10(ISBN_10 uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Book{}).Where("books.isbn_10 = ?", ISBN_10).Count(&count).Error
	exists = count > 0
	return
}

func (m *BookModel) DoesBookExistByISBN13(ISBN_13 uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Book{}).Where("books.isbn_13 = ?", ISBN_13).Count(&count).Error
	exists = count > 0
	return
}

func (m *BookModel) CreateBook(b *Book) (err error) {
	err = m.DB.FirstOrCreate(b).Error
	if err != nil {
		return err
	}

	err = m.DB.Find(b).First(b).Error
	return
}

func (m *BookModel) AddCoursesToBook(id uint, courseIDs *[]uint) (err error) {
	book := new(Book)
	err = m.DB.First(&book, id).Error
	if err != nil {
		return err
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
				return lib.ErrorBookCourseNotFound
			}
			return err
		}

		err = tx.Model(&book).Association("Courses").Append(course).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return
	}
	return
}

func (m *BookModel) DeleteBook(b *Book) (err error) {
	err = m.DB.Delete(b).Error
	return
}

func checkValidISBN(ISBN string, length int) bool {
	if len(ISBN) != length {
		return false
	}
	if _, err := strconv.ParseUint(ISBN, 10, 64); err != nil {
		return false
	}
	return true
}
