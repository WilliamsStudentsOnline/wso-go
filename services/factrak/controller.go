package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	professorModel *models.ProfessorModel
	departmentModel *models.Department
	courseModel *models.CourseModel
	agreementModel *models.FactrakAgreementModel
	surveyModel *models.FactrakSurveyModel
}

// Construct a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		professorModel: &models.ProfessorModel{
			UserModel: &models.UserModel{
				BaseModel: models.BaseModel{
					DB: db,
				},
			},
		},
	}
}