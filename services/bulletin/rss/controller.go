package rss

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	threads *models.BoardThreadModel
	cfg     *config.Config
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		threads:        models.NewBoardThreadModel(db, log),
		cfg:            cfg,
	}
}

// ListLostAndFoundBulletins godoc
// @Summary List lost and found bulletins RSS
// @Description lists lost and found board threads in RSS
// @ID listLostAndFoundRss
// @Tags bulletins
// @Produce xml
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/lostAndFound [get]
func (t *Controller) ListLostAndFoundBulletins(c *gin.Context) {
	t.listThreadsByType(c, models.BoardThreadTypeLostAndFound, "WSO Bulletins lostAndFound")
}

// ListJobBulletins godoc
// @Summary List job bulletins RSS
// @Description lists job board threads in RSS
// @ID listJobRss
// @Tags bulletins
// @Produce xml
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/job [get]
func (t *Controller) ListJobBulletins(c *gin.Context) {
	t.listThreadsByType(c, models.BoardThreadTypeJob, "WSO Bulletins job")
}

// ListExchangeBulletins godoc
// @Summary List exchange bulletins RSS
// @Description lists exchange board threads in RSS
// @ID listExchangeRss
// @Tags bulletins
// @Produce xml
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/exchange [get]
func (t *Controller) ListExchangeBulletins(c *gin.Context) {
	t.listThreadsByType(c, models.BoardThreadTypeExchange, "WSO Bulletins exchange")
}

// ListAnnouncementBulletins godoc
// @Summary List announcement bulletins RSS
// @Description lists announcement board threads in RSS
// @ID listAnnouncementRss
// @Tags bulletins
// @Produce xml
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/announcement [get]
func (t *Controller) ListAnnouncementBulletins(c *gin.Context) {
	t.listThreadsByType(c, models.BoardThreadTypeAnnouncement, "WSO Bulletins announcement")
}

// ListRideBulletins godoc
// @Summary List ride bulletins RSS
// @Description lists ride board threads in RSS
// @ID listRideRss
// @Tags bulletins
// @Produce xml
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param offeringRide query bool false "Filter offering vs requesting"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/ride [get]
func (t *Controller) ListRideBulletins(c *gin.Context) {
	opts := models.GetBoardThreadsOptions{
		Types: []string{models.BoardThreadTypeRide},
		Sort:  "created",
	}
	if err := c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	opts.Types = []string{models.BoardThreadTypeRide}
	opts.Sort = "created"

	var threads []*models.BoardThread
	if err := t.threads.GetThreads(&threads, &opts); err != nil {
		t.RespondError(c, err)
		return
	}

	segment := feedPathSegment(models.BoardThreadTypeRide)
	baseURL := t.cfg.GenerateURL()
	baseURL.Path = "/bulletin/" + segment
	feed := &feeds.Feed{
		Title:   "WSO Ride Bulletins",
		Created: time.Now(),
		Items:   make([]*feeds.Item, len(threads)),
		Link:    &feeds.Link{Href: baseURL.String()},
	}

	for i, thread := range threads {
		strID := strconv.Itoa(int(thread.ID))
		itemURL := t.cfg.GenerateURL()
		itemURL.Path = "/bulletin/" + segment + "/" + strID

		title := thread.Title
		created := thread.CreatedAt
		if thread.RideInformation != nil {
			rideType := "request"
			if thread.RideInformation.OfferingRide != nil && *thread.RideInformation.OfferingRide {
				rideType = "offer"
			}
			title = fmt.Sprintf("%s to %s (%s)",
				thread.RideInformation.Source,
				thread.RideInformation.Destination,
				rideType,
			)
			if thread.StartsAt != nil {
				created = *thread.StartsAt
			}
		}

		feed.Items[i] = &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: itemURL.String()},
			Author:      &feeds.Author{Name: threadAuthorName(thread)},
			Id:          strID,
			Created:     created,
			Description: threadBody(thread),
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
// @Description lists discussion board threads in RSS
// @ID listDiscussionRss
// @Tags bulletins
// @Produce xml
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Success 200 {string} string "A RSS Feed serialized as XML string"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /bulletin/rss/discussion [get]
func (t *Controller) ListDiscussionBulletins(c *gin.Context) {
	t.listThreadsByType(c, models.BoardThreadTypeDiscussion, "WSO Discussions")
}

func (t *Controller) listThreadsByType(c *gin.Context, threadType, title string) {
	opts := models.GetBoardThreadsOptions{
		Types: []string{threadType},
		Sort:  "created",
	}
	if err := c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	opts.Types = []string{threadType}
	opts.Sort = "created"

	var threads []*models.BoardThread
	if err := t.threads.GetThreads(&threads, &opts); err != nil {
		t.RespondError(c, err)
		return
	}

	segment := feedPathSegment(threadType)
	baseURL := t.cfg.GenerateURL()
	baseURL.Path = "/bulletin/" + segment
	feed := &feeds.Feed{
		Title:   title,
		Created: time.Now(),
		Items:   make([]*feeds.Item, len(threads)),
		Link:    &feeds.Link{Href: baseURL.String()},
	}

	for i, thread := range threads {
		strID := strconv.Itoa(int(thread.ID))
		itemURL := t.cfg.GenerateURL()
		itemURL.Path = "/bulletin/" + segment + "/" + strID
		feed.Items[i] = &feeds.Item{
			Title:       thread.Title,
			Link:        &feeds.Link{Href: itemURL.String()},
			Author:      &feeds.Author{Name: threadAuthorName(thread)},
			Id:          strID,
			Updated:     thread.UpdatedAt,
			Created:     thread.CreatedAt,
			Description: threadBody(thread),
		}
	}

	res, err := feed.ToRss()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	c.String(http.StatusOK, res)
}

func feedPathSegment(threadType string) string {
	if threadType == models.BoardThreadTypeDiscussion {
		return "discussions"
	}
	return threadType
}

func threadBody(thread *models.BoardThread) string {
	for _, p := range thread.Posts {
		if p != nil && p.IsOP != nil && *p.IsOP {
			return p.Content
		}
	}
	if len(thread.Posts) > 0 && thread.Posts[0] != nil {
		return thread.Posts[0].Content
	}
	return thread.Body
}

func threadAuthorName(thread *models.BoardThread) string {
	name := userToAnonName(thread.User)
	if name == "Unknown" && thread.ExUserName != "" {
		return nameToAnonName(thread.ExUserName)
	}
	return name
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
