package banner

import (
	"encoding/json"
	"errors"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var client *redis_util.RedisClient

func ConfigureClient(addr string) error {
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	var err error
	client, err = redis_util.SetupClient(addr, "", redis_util.BannerDatabaseID)
	return err
}

func ConfigureControllerForTest() error {
	var err error
	client, err = redis_util.SetupRedisClientForTest()
	return err
}

type Controller struct {
	services.BaseController
}

// Banner types from https://github.com/WilliamsStudentsOnline/wso-go/issues/81
const (
	BannerTypeEmergency   = "emergency"
	BannerTypePinned      = "pinned"
	BannerTypeDismissible = "dismissible"
)

// Banner matches the homepage join-header fields in wso-react Homepage.tsx.
// Visible controls whether the banner is shown; type/text/link/sub describe the notice.
type Banner struct {
	Visible bool   `json:"visible"`
	Type    string `json:"type" binding:"omitempty,oneof=emergency pinned dismissible"`
	Text    string `json:"text"`
	Link    string `json:"link"`
	Sub     string `json:"sub"`
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
	}
}

// GetBanner godoc
// @Summary Get homepage banner
// @Description get the current homepage join-header banner (public)
// @ID getBanner
// @Tags banner
// @Accept json
// @Produce json
// @Success 200 {object} banner.Banner
// @Failure 500 {object} services.BaseErrorResponse
// @Router /banner [get]
func (t *Controller) GetBanner(c *gin.Context) {
	val, err := client.Get(c, redis_util.GetHomepageBannerKey())
	if err == redis_util.ErrRedisClientNotConfigured {
		t.RespondAPIError(c, lib.ErrorRedisClientNotConfigured)
		return
	}
	if err == redis.Nil {
		t.RespondOK(c, Banner{})
		return
	}
	if err != nil {
		t.RespondError(c, err)
		return
	}

	valStr, ok := val.(string)
	if !ok {
		t.RespondAPIError(c, lib.ErrorBannerUnmarshal)
		return
	}

	var result Banner
	if err := json.Unmarshal([]byte(valStr), &result); err != nil {
		t.RespondAPIError(c, lib.ErrorBannerUnmarshal)
		return
	}

	t.RespondOK(c, result)
}

// SetBanner godoc
// @Summary Set homepage banner
// @Description set the homepage join-header banner (admin only). Set visible=false to hide.
// @ID setBanner
// @Tags banner
// @Accept json
// @Produce json
// @Param banner body banner.Banner true "Banner fields"
// @Success 200 {object} banner.Banner
// @Failure 403 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /banner [post]
func (t *Controller) SetBanner(c *gin.Context) {
	var body Banner
	if err := c.ShouldBindJSON(&body); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	if body.Visible && body.Type == "" {
		t.RespondBadBind(c, errors.New("type is required when visible is true"))
		return
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	err = client.Set(c, redis_util.GetHomepageBannerKey(), string(payload))
	if err == redis_util.ErrRedisClientNotConfigured {
		t.RespondAPIError(c, lib.ErrorRedisClientNotConfigured)
		return
	}
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, body)
}
