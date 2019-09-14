package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	search "github.com/WilliamsStudentsOnline/wso-go/lib/search/factrak"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	// Put models here:
	professorModel   *models.ProfessorModel
	userModel        *models.UserModel
	courseModel      *models.CourseModel
	departmentModel  *models.DepartmentModel
	areaOfStudyModel *models.AreaOfStudyModel
	agreementModel   *models.FactrakAgreementModel
	surveyModel      *models.FactrakSurveyModel
	studentModel     *models.StudentModel
	factrakSearch    search.SearchFactrak
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config) *Controller {
	return &Controller{
		professorModel:   models.NewProfessorModel(db),
		userModel:        models.NewUserModel(db),
		courseModel:      models.NewCourseModel(db),
		departmentModel:  models.NewDepartmentModel(db),
		areaOfStudyModel: models.NewAreaOfStudyModel(db),
		agreementModel:   models.NewFactrakAgreementModel(db),
		surveyModel:      models.NewFactrakSurveyModel(db),
		studentModel:     models.NewStudentModel(db),
		factrakSearch:    search.NewSearchFactrak(db, cfg),
	}
}

// Remove sensitive data, like userID from surveys. Unless scope admin or the survey is your own
func RemoveUserIDFromSurveys(c *gin.Context, s []*models.FactrakSurvey) {
	if auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeFactrakAdmin) {
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

func IsScopeLimited(c *gin.Context) bool {
	return !auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeFactrakAdmin, auth.ScopeFactrakFull)
}

func IsScopeAdmin(c *gin.Context) bool {
	return auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeFactrakAdmin)
}

// Get specified query ID (eg courseID). If it does not exist, return nil. If there is an error, abort.
// Be sure to check if the context has been aborted after calling this.
func (t *Controller) getQueryID(c *gin.Context, key string) *uint {
	// If there is no query ID, return nil
	if _, queryIDExists := c.GetQuery(key); !queryIDExists {
		return nil
	}

	// If there is a query ID, look at that.
	queryID, err := services.GetUIntQuery(c, key)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return nil
	}

	return &queryID
}
