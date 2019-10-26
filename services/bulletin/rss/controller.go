package rss

import (
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services/bulletin"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	*bulletin.Controller
}

// NewController constructs a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		Controller: bulletin.NewController(db),
	}
}

type Link struct {
	Href, Rel, Type, Length string
}

type Author struct {
	Name string
}

type Enclosure struct {
	Url, Length, Type string
}

type Item struct {
	Title       string
	Link        *Link
	Source      *Link
	Author      *Author
	Description string // used as description in rss, summary in atom
	Id          string // used as guid in rss, id in atom
	Updated     time.Time
	Created     time.Time
	Enclosure   *Enclosure
	Content     string
}

type Feed struct {
	Title       string
	Link        *Link
	Description string
	Author      *Author
	Updated     time.Time
	Created     time.Time
	Id          string
	Subtitle    string
	Items       []*Item
	Copyright   string
}

// ListBulletinsRSS godoc
// @Summary List bulletins RSS
// @Description lists all bulletins in RSS
// @ID bulletins-list-bulletins-rss
// @Tags bulletins
// @Produce  xml
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param type query string false "Bulletin Type"
// @Param all query string false "Get All Bulletins (no restriction on startDate, endDate)"
// @Success 200 {array} models.Bulletin
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/bulletins [get]
func (t *Controller) ListBulletinsRSS(c *gin.Context) {
	params := new(models.GetAllBulletinsOptions)

	err := c.ShouldBindQuery(params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	var bulletins []*models.Bulletin
	err = t.bulletinModel.GetAllBulletinsWithOptions(&bulletins, params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.bulletinModel.CountAllBulletinsWithOptions(params)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	// Remove user info if not a user. Need this, as bulletin service is public
	removeUserInfoFromBulletin(c, bulletins)

	t.RespondOK(c, bulletins)
}
