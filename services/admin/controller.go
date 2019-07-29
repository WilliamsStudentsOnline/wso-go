package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	userModel    *models.UserModel
	studentModel *models.StudentModel
	cfg          *config.Config
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config) *Controller {
	return &Controller{
		userModel:    models.NewUserModel(db),
		studentModel: models.NewStudentModel(db),
		cfg:          cfg,
	}
}

// Calls the UpdateAllFromLDAP from Users
// TODO: This endpoint runs very slow; may want to return a job ID and then be able to query the job log (Aidan: like I did for brkt)
// TODO: If we do have a job setup, will need to make it work with multiple deployments/scaling
func (t *Controller) UpdateAllUsersFromLDAP(c *gin.Context) {
	err := t.userModel.UpdateAllFromLDAP(t.cfg)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}

// Calls the UpdateAllFactrakSurveyDeficits from Students
// TODO: This endpoint runs very slow; may want to return a job ID and then be able to query the job log (Aidan: like I did for brkt)
// TODO: If we do have a job setup, will need to make it work with multiple deployments/scaling
func (t *Controller) UpdateAllFactrakSurveyDeficits(c *gin.Context) {
	err := t.studentModel.UpdateAllFactrakSurveyDeficits()

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
