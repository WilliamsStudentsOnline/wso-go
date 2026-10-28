package feedback

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	feedbackModel *models.FeedbackModel
	cfg           *config.Config
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		feedbackModel:  models.NewFeedbackModel(db, log),
		cfg:            cfg,
	}
}

type CreateFeedbackParams struct {
	UserID        *uint   `json:"userID"`
	PreferredName *string `json:"preferredName"`
	URL           *string `json:"url"`
	Notes         string  `json:"notes" binding:"required"`
}

type UpdateFeedbackParams struct {
	AdminNotes *string                  `json:"adminNotes"`
	Priority   *models.FeedbackPriority `json:"priority"`
	Issue      *string                  `json:"issue"`
}
