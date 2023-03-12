package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
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
