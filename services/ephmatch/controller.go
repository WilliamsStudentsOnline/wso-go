package ephmatch

import (
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	ephmatchModel *models.EphmatchModel
	profileModel  *models.EphmatchProfileModel
	likeModel     *models.EphmatchLikeModel
	matchModel    *models.EphmatchMatchesModel
	cfg           *config.Config
}

// Construct a new dormtrak controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		ephmatchModel:  models.NewEphmatchModel(db, log),
		profileModel:   models.NewEphmatchProfileModel(db, log),
		likeModel:      models.NewEphmatchLikeModel(db, log),
		matchModel:     models.NewEphmatchMatchesModel(db, log),
		cfg:            cfg,
	}
}

type GetAvailabilityResp struct {
	Available        bool       `json:"available"`        // If Ephmatch is currently available
	OpenIndefinitely bool       `json:"openIndefinitely"` // If Ephmatch has no closing time set
	ClosingTime      *time.Time `json:"closingTime"`      // Closing time for current Ephmatch era/period
	NextOpenTime     *time.Time `json:"nextOpenTime"`     // Next time Ephmatch will be open
}

// GetAvailability godoc
// @Summary Get availability
// @Description gives data on Ephmatch availability
// @ID ephmatch-get-availability
// @Tags ephmatch
// @Produce  json
// @Success 200 {object} GetAvailabilityResp
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/availability [get]
func (t *Controller) GetAvailability(c *gin.Context) {
	resp := GetAvailabilityResp{}
	now := time.Now()

	// Fill in data on availability and closing time
	if t.cfg.EphmatchEnableNow {
		resp.Available = true
		resp.OpenIndefinitely = true
	} else {
		for _, era := range t.cfg.EphmatchEras {
			if era.Start.Before(now) && era.End.After(now) {
				resp.Available = true
				resp.ClosingTime = &era.End
				break
			}
		}
	}

	// Fill in next open time
	var nextEraStart *time.Time
	for _, era := range t.cfg.EphmatchEras {
		// If era is after now and before current closest era, it is out nextEraStart
		if era.Start.After(now) {
			if nextEraStart == nil {
				nextEraStart = &era.Start
			} else if era.Start.Before(*nextEraStart) {
				nextEraStart = &era.Start
			}
		}
	}
	resp.NextOpenTime = nextEraStart

	t.RespondOK(c, resp)
}
