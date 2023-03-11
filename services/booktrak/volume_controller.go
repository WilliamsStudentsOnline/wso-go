package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/books/v1"
)

type SearchBooksRequest struct {
	Query string `form:"q"`
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

	t.RespondOK(c, volumes)
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

func bookMatchesOnlineData(book models.Book, onlineBook *books.VolumeVolumeInfo) bool {
	if book.Title != onlineBook.Title {
		return false
	}
	if book.Subtitle != onlineBook.Subtitle {
		return false
	}
	if !lib.StringSlicesEqual(book.Authors, onlineBook.Authors) {
		return false
	}
	if book.Publisher != onlineBook.Publisher {
		return false
	}

	// Might need to change if isbn 10 isn't included in every book
	for _, v := range onlineBook.IndustryIdentifiers {
		if v.Type == "ISBN_10" {
			if v.Identifier != book.ISBN_10 {
				return false
			}
		}
		if v.Type == "ISBN_13" {
			if v.Identifier != book.ISBN_13 {
				return false
			}
		}
	}

	if book.InfoLink != onlineBook.InfoLink {
		return false
	}
	if book.ImageLink != onlineBook.ImageLinks.Thumbnail {
		return false
	}

	return true
}
