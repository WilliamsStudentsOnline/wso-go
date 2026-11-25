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
	db                 *gorm.DB
	userModel          *models.UserModel
	bannedUserModel    *models.BannedUserModel
	studentModel       *models.StudentModel
	tagModel           *models.TagModel
	moderatorTaskModel *models.ModeratorTaskModel
	surveyModel        *models.FactrakSurveyModel
	jobRunModel        *models.JobRunModel
	boardThreadModel   *models.BoardThreadModel
	boardPostModel     *models.BoardPostModel
	exemptionModel     *models.FactrakDeficitExemptionModel
	cfg                *config.Config
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:     services.BaseController{Log: log},
		db:                 db,
		userModel:          models.NewUserModel(db, log),
		bannedUserModel:    models.NewBannedUserModel(db, log),
		studentModel:       models.NewStudentModel(db, log),
		tagModel:           models.NewTagModel(db, log),
		moderatorTaskModel: models.NewModeratorTaskModel(db, log),
		surveyModel:        models.NewFactrakSurveyModel(db, log),
		jobRunModel:        models.NewJobRunModel(db, log),
		boardThreadModel:   models.NewBoardThreadModel(db, log),
		boardPostModel:     models.NewBoardPostModel(db, log),
		exemptionModel:     models.NewFactrakDeficitExemptionModel(db, log),
		cfg:                cfg,
	}
}
