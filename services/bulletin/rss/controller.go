package rss

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	bulletinModel   *models.BulletinModel
	rideModel       *models.BulletinRideModel
	discussionModel *models.DiscussionModel
	cfg             *config.Config
}

// NewController constructs a new user controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:  services.BaseController{Log: log},
		bulletinModel:   models.NewBulletinModel(db, log),
		rideModel:       models.NewBulletinRideModel(db, log),
		discussionModel: models.NewDiscussionModel(db, log),
		cfg:             cfg,
	}
}

// ListLostAndFoundBulletins godoc
// @Summary List lost and found bulletins RSS
// @Description lists lost and found bulletins in RSS
// @ID listLostAndFoundRss
// @Tags bulletins
// @Produce xml
// @Param start query string false "Start Pagination (timestamp)"
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param all query string false "Get All Bulletins (no restriction on startDate, endDate)"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/lostAndFound [get]
func (t *Controller) ListLostAndFoundBulletins(c *gin.Context) {
	t.doListGeneralBulletinsByType(c, models.BulletinTypeLostAndFound)
}

// ListJobBulletins godoc
// @Summary List job bulletins RSS
// @Description lists job bulletins in RSS
// @ID listJobRss
// @Tags bulletins
// @Produce xml
// @Param start query string false "Start Pagination (timestamp)"
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param all query string false "Get All Bulletins (no restriction on startDate, endDate)"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/job [get]
func (t *Controller) ListJobBulletins(c *gin.Context) {
	t.doListGeneralBulletinsByType(c, models.BulletinTypeJob)
}

// ListExchangeBulletins godoc
// @Summary List exchange bulletins RSS
// @Description lists exchange bulletins in RSS
// @ID listExchangeRss
// @Tags bulletins
// @Produce xml
// @Param start query string false "Start Pagination (timestamp)"
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param all query string false "Get All Bulletins (no restriction on startDate, endDate)"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/exchange [get]
func (t *Controller) ListExchangeBulletins(c *gin.Context) {
	t.doListGeneralBulletinsByType(c, models.BulletinTypeExchange)
}

// ListAnnouncementBulletins godoc
// @Summary List announcement bulletins RSS
// @Description lists announcement bulletins in RSS
// @ID listAnnouncementRss
// @Tags bulletins
// @Produce xml
// @Param start query string false "Start Pagination (timestamp)"
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param all query string false "Get All Bulletins (no restriction on startDate, endDate)"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/announcement [get]
func (t *Controller) ListAnnouncementBulletins(c *gin.Context) {
	t.doListGeneralBulletinsByType(c, models.BulletinTypeAnnouncement)
}

// ListRideBulletins godoc
// @Summary List ride bulletins RSS
// @Description lists ride bulletins in RSS
// @ID listRideRss
// @Tags bulletins
// @Produce xml
// @Param start query string false "Start Pagination (timestamp)"
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param type query string false "Ride Type (request, offer)"
// @Param all query string false "Get All Bulletins (no restriction on startDate, endDate)"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/ride [get]
func (t *Controller) ListRideBulletins(c *gin.Context) {
	params := models.GetAllBulletinRidesOptions{}

	err := c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	params.Preload = []string{"user"}

	var rides []*models.BulletinRide
	err = t.rideModel.GetAllRides(&rides, &params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	baseURL := t.cfg.GenerateURL()
	baseURL.Path = "/bulletins/ride"
	feed := &feeds.Feed{
		Title:   "WSO Ride Bulletins",
		Created: time.Now(),
		Items:   make([]*feeds.Item, len(rides)),
		Link: &feeds.Link{
			Href: baseURL.String(),
		},
	}

	for i := range rides {
		strID := strconv.Itoa(int(rides[i].ID))

		itemURL := t.cfg.GenerateURL()
		itemURL.Path = "/bulletins/ride/" + strID

		rideType := "unknown"
		if rides[i].Offer != nil {
			if *rides[i].Offer {
				rideType = "offer"
			} else {
				rideType = "request"
			}
		}

		feed.Items[i] = &feeds.Item{
			Title: fmt.Sprintf("%s to %s (%s)", rides[i].Source, rides[i].Destination, rideType),
			Link: &feeds.Link{
				Href: itemURL.String(),
			},
			Author: &feeds.Author{
				Name: userToAnonName(rides[i].User),
			},
			Id:          strID,
			Created:     rides[i].Date,
			Description: rides[i].Body,
		}
	}

	res, err := feed.ToRss()
	if err != nil {
		t.RespondError(c, err)
		return
	}

	c.String(http.StatusOK, res)
}

// ListDiscussionBulletins godoc
// @Summary List discussion bulletins RSS
// @Description lists discussion bulletins in RSS
// @ID listDiscussionRss
// @Tags bulletins
// @Produce xml
// @Param start query string false "Start Pagination (timestamp)"
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/discussion [get]
func (t *Controller) ListDiscussionBulletins(c *gin.Context) {
	params := models.GetAllDiscussionsOptions{}

	err := c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	params.Preload = []string{"user"}
	params.GetLastPost = lib.TruePtr()

	var discussions []*models.Discussion
	err = t.discussionModel.GetAllDiscussions(&discussions, &params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	baseURL := t.cfg.GenerateURL()
	baseURL.Path = "/bulletins/discussions"
	feed := &feeds.Feed{
		Title:   "WSO Discussions",
		Created: time.Now(),
		Items:   make([]*feeds.Item, len(discussions)),
		Link: &feeds.Link{
			Href: baseURL.String(),
		},
	}

	for i := range discussions {
		strID := strconv.Itoa(int(discussions[i].ID))

		itemURL := t.cfg.GenerateURL()
		itemURL.Path = "/bulletins/discussions/threads/" + strID

		name := userToAnonName(discussions[i].User)
		if name == "Unknown" && discussions[i].ExUserName != "" {
			name = nameToAnonName(discussions[i].ExUserName)
		}

		firstPost := ""
		if len(discussions[i].Posts) > 0 {
			firstPost = discussions[i].Posts[0].Content
		}

		feed.Items[i] = &feeds.Item{
			Title: discussions[i].Title,
			Link: &feeds.Link{
				Href: itemURL.String(),
			},
			Author: &feeds.Author{
				Name: name,
			},
			Id:          strID,
			Created:     discussions[i].CreatedAt,
			Updated:     discussions[i].UpdatedAt,
			Description: firstPost,
		}
	}

	res, err := feed.ToRss()
	if err != nil {
		t.RespondError(c, err)
		return
	}

	c.String(http.StatusOK, res)
}

func (t *Controller) doListGeneralBulletinsByType(c *gin.Context, bulletinType string) {
	params := new(models.GetAllBulletinsOptions)

	err := c.ShouldBindQuery(params)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	params.Preload = []string{"user"}
	params.Type = &bulletinType

	var bulletins []*models.Bulletin
	err = t.bulletinModel.GetAllBulletinsWithOptions(&bulletins, params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	baseURL := t.cfg.GenerateURL()
	baseURL.Path = "/bulletins/" + bulletinType
	feed := &feeds.Feed{
		Title:   "WSO Bulletins " + bulletinType,
		Created: time.Now(),
		Items:   make([]*feeds.Item, len(bulletins)),
		Link: &feeds.Link{
			Href: baseURL.String(),
		},
	}

	for i := range bulletins {
		strID := strconv.Itoa(int(bulletins[i].ID))

		itemURL := t.cfg.GenerateURL()
		itemURL.Path = "/bulletins/" + bulletinType + "/" + strID
		feed.Items[i] = &feeds.Item{
			Title: bulletins[i].Title,
			Link: &feeds.Link{
				Href: itemURL.String(),
			},
			Author: &feeds.Author{
				Name: userToAnonName(bulletins[i].User),
			},
			Id:          strID,
			Updated:     bulletins[i].UpdatedAt,
			Created:     bulletins[i].CreatedAt,
			Description: bulletins[i].Body,
		}
	}

	res, err := feed.ToRss()
	if err != nil {
		t.RespondError(c, err)
		return
	}

	c.String(http.StatusOK, res)
}

func userToAnonName(user *models.User) string {
	if user == nil {
		return "Unknown"
	}

	return nameToAnonName(user.Name)
}

func nameToAnonName(name string) string {
	nameSplit := strings.Split(name, " ")

	anonName := nameSplit[0]

	if len(nameSplit) > 1 {
		anonName += " " + string(nameSplit[len(nameSplit)-1][0])
	}

	return anonName
}
