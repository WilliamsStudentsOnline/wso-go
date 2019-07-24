package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	// Put models here:
	professorModel *models.ProfessorModel
	userModel *models.UserModel
	departmentModel *models.DepartmentModel
	courseModel *models.CourseModel
	agreementModel *models.FactrakAgreementModel
	surveyModel *models.FactrakSurveyModel
}

// Construct a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		professorModel: models.NewProfessorModel(db),
		userModel: models.NewUserModel(db),
		departmentModel: models.NewDepartmentModel(db),
		courseModel: models.NewCourseModel(db),
		agreementModel: models.NewFactrakAgreementModel(db),
		surveyModel: models.NewFactrakSurveyModel(db),
	}
}

// Remove sensitive data, like userID from surveys. Unless scope admin or the survey is your own
func RemoveUserIDFromSurveys(c *gin.Context, s []*models.FactrakSurvey) {
	if auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeAdminFactrak) {
		return
	}

	userID := services.GetUserID(c)

	for _, survey := range s {
		if survey.UserID == userID {
			continue
		}
		survey.UserID = 0
		survey.User = nil
	}
}