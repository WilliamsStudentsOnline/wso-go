package ephmatch

import (
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	ephmatchModel  *models.EphmatchModel
	profileModel   *models.EphmatchProfileModel
	relationModel  *models.EphmatchRelationModel
	matchModel     *models.EphmatchMatchesModel
	cfg            *config.Config
	pictureBackend pictures.PictureBackend
}

// Construct a new dormtrak controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	pb, err := pictures.NewPictureBackend(cfg, log)
	if err != nil {
		log.Error(err)
		log.Warn("Using picture backend none")
		pb = pictures.NewPictureBackendDummy()
	}

	return &Controller{
		BaseController: services.BaseController{Log: log},
		ephmatchModel:  models.NewEphmatchModel(db, log),
		profileModel:   models.NewEphmatchProfileModel(db, log),
		relationModel:  models.NewEphmatchRelationModel(db, log),
		matchModel:     models.NewEphmatchMatchesModel(db, log),
		cfg:            cfg,
		pictureBackend: pb,
	}
}

type GetAvailabilityResp struct {
	Available        bool       `json:"available"`        // If Ephmatch is currently available
	OpenIndefinitely bool       `json:"openIndefinitely"` // If Ephmatch has no closing time set
	ClosingTime      *time.Time `json:"closingTime"`      // Closing time for current Ephmatch era/period
	NextOpenTime     *time.Time `json:"nextOpenTime"`     // Next time Ephmatch will be open
	SeniorOnly       bool       `json:"seniorOnly"`       // Senior only ephmatch right now
}

// GetAvailability godoc
// @Summary Get availability
// @Description gives data on Ephmatch availability
// @ID ephmatch-get-availability
// @Tags ephmatch
// @Produce  json
// @Success 200 {object} ephmatch.GetAvailabilityResp
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /ephmatch/availability [get]
func (t *Controller) GetAvailability(c *gin.Context) {
	resp := GetAvailabilityResp{}
	now := time.Now()

	// Fill in data on availability and closing time
	if t.cfg.EphmatchEnableNow {
		resp.Available = true
		resp.OpenIndefinitely = true
		resp.SeniorOnly = false
	} else {
		for _, era := range t.cfg.EphmatchEras {
			if era.Start.Before(now) && era.End.After(now) {
				resp.Available = true
				resp.ClosingTime = &era.End
				resp.SeniorOnly = era.SeniorOnly
				break
			}
		}
	}

	// Fill in next open time
	var nextEraStart *time.Time
	nextEraSeniorOnly := false
	for _, era := range t.cfg.EphmatchEras {
		// If era is after now and before current closest era, it is out nextEraStart
		if era.Start.After(now) {
			if nextEraStart == nil {
				// We need to copy era.Start to a new object, instead of pointing to it (which changes during iteration)
				startLocal := era.Start
				// Note that dangling pointer problems do not exist in golang
				nextEraStart = &startLocal
				nextEraSeniorOnly = era.SeniorOnly
			} else if era.Start.Before(*nextEraStart) {
				*nextEraStart = era.Start
				nextEraSeniorOnly = era.SeniorOnly
			}
		}
	}
	resp.NextOpenTime = nextEraStart

	if !resp.Available {
		resp.SeniorOnly = nextEraSeniorOnly
	}

	t.RespondOK(c, resp)
}

func (t *Controller) isSeniorOnly() bool {
	now := time.Now()

	if !t.cfg.EphmatchEnableNow {
		for _, era := range t.cfg.EphmatchEras {
			if era.Start.Before(now) && era.End.After(now) {
				return era.SeniorOnly
			}
		}
	}

	return false
}
