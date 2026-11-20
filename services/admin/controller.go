package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	db                 *gorm.DB
	userModel          *models.UserModel
	studentModel       *models.StudentModel
	tagModel           *models.TagModel
	moderatorTaskModel *models.ModeratorTaskModel
	surveyModel        *models.FactrakSurveyModel
	jobRunModel        *models.JobRunModel
	boardThreadModel   *models.BoardThreadModel
	boardPostModel     *models.BoardPostModel
	pictureBackend     pictures.PictureBackend
	cfg                *config.Config
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	pb, err := pictures.NewPictureBackend(cfg, log)
	if err != nil {
		log.Error(err)
		log.Warn("Using picture backend none")
		pb = pictures.NewPictureBackendDummy()
	}

	return &Controller{
		BaseController:     services.BaseController{Log: log},
		db:                 db,
		userModel:          models.NewUserModel(db, log),
		studentModel:       models.NewStudentModel(db, log),
		tagModel:           models.NewTagModel(db, log),
		moderatorTaskModel: models.NewModeratorTaskModel(db, log),
		surveyModel:        models.NewFactrakSurveyModel(db, log),
		jobRunModel:        models.NewJobRunModel(db, log),
		boardThreadModel:   models.NewBoardThreadModel(db, log),
		boardPostModel:     models.NewBoardPostModel(db, log),
		pictureBackend:     pb,
		cfg:                cfg,
	}
}
