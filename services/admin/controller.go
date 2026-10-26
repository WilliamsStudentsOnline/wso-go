package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	userModel          *models.UserModel
	studentModel       *models.StudentModel
	tagModel           *models.TagModel
	moderatorTaskModel *models.ModeratorTaskModel
	surveyModel        *models.FactrakSurveyModel
	cfg                *config.Config
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:     services.BaseController{Log: log},
		userModel:          models.NewUserModel(db, log),
		studentModel:       models.NewStudentModel(db, log),
		tagModel:           models.NewTagModel(db, log),
		moderatorTaskModel: models.NewModeratorTaskModel(db, log),
		surveyModel:        models.NewFactrakSurveyModel(db, log),
		cfg:                cfg,
	}
}
