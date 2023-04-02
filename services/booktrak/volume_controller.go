package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/books/v1"
)

type SearchBooksRequest struct {
	Query string `form:"q" binding:"required"`
	Limit *int   `form:"limit"`
}

func (t *Controller) SearchBooks(c *gin.Context) {
	var err error
	request := SearchBooksRequest{}
	if err = c.ShouldBindQuery(&request); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	volumes, err := t.searchVolumes(request.Query, request.Limit)

	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}

	t.RespondOK(c, *volumesToBooks(&volumes.Items))
}

func volumesToBooks(volumes *[]*books.Volume) *[]*models.Book {
	books := make([]*models.Book, 0, len(*volumes))
	for _, volume := range *volumes {
		isbn10, isbn13 := "", ""
		for _, v := range volume.VolumeInfo.IndustryIdentifiers {
			if v.Type == "ISBN_10" {
				isbn10 = v.Identifier
			}
			if v.Type == "ISBN_13" {
				isbn13 = v.Identifier
			}
		}

		book := models.Book{
			Title:     volume.VolumeInfo.Title,
			Subtitle:  volume.VolumeInfo.Subtitle,
			Authors:   volume.VolumeInfo.Authors,
			Publisher: volume.VolumeInfo.Publisher,
			InfoLink:  volume.VolumeInfo.InfoLink,
			ISBN_10:   isbn10,
			ISBN_13:   isbn13,
		}
		if volume.VolumeInfo.ImageLinks != nil {
			book.ImageLink = volume.VolumeInfo.ImageLinks.Thumbnail
		}
		books = append(books, &book)
	}

	return &books
}

func (t *Controller) searchVolumes(query string, limit *int) (volumes *books.Volumes, err error) {

	if limit == nil || *limit < 0 {
		limit = lib.IntToPtr(10)
	}

	volumes, err = t.volumeService.
		List(query).
		MaxResults(int64(*limit)).
		Fields("items/volumeInfo/authors",
			"items/volumeInfo/imageLinks",
			"items/volumeInfo/industryIdentifiers",
			"items/volumeInfo/infoLink",
			"items/volumeInfo/title",
			"items/volumeInfo/subtitle",
			"items/volumeInfo/publisher").
		Do()

	return
}
