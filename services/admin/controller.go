package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/jobs"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	v1 "k8s.io/api/batch/v1"
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

type KubeJobReturn struct {
	JobID string `json:"jobID"`
}

// Updates the course catalog
// CatalogUpdate godoc
// @Summary Updates the course catalog
// @Description executes a kubernetes job that updates the course catalog
// @ID catalog-update
// @Tags admin
// @Accept  json
// @Produce  json
// @Success 201 {object} admin.KubeJobReturn
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /admin/catalog-update [post]
func (t *Controller) CatalogUpdate(c *gin.Context) {
	var job *v1.Job
	var err error
	var resp KubeJobReturn

	if t.cfg.KubernetesEnabled {
		job, err = jobs.RunCatalogUpdateJob(t.cfg)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	if job != nil {
		resp = KubeJobReturn{
			JobID: job.Name,
		}
	}

	t.RespondCreated(c, resp)
}

// Calls the UpdateAllFromLDAP from Users
// TODO: This endpoint runs very slow; may want to return a job ID and then be able to query the job log (Aidan: like I did for brkt)
// TODO: If we do have a job setup, will need to make it work with multiple deployments/scaling
// UpdateAllUsersFromLDAP godoc
// @Summary Updates all users from LDAP
// @Description runs a slow job that updates all users from LDAP
// @ID update-all-users-from-ldap
// @Tags admin
// @Accept  json
// @Produce  json
// @Success 201 {object} admin.KubeJobReturn
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /admin/update-all-users-from-ldap [post]
func (t *Controller) UpdateAllUsersFromLDAP(c *gin.Context) {
	var job *v1.Job
	var err error
	var resp KubeJobReturn

	if t.cfg.KubernetesEnabled {
		job, err = jobs.RunUpdateAllUsersFromLDAPJob(t.cfg)
	} else {
		err = t.userModel.UpdateAllFromLDAP(t.cfg)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	if job != nil {
		resp = KubeJobReturn{
			JobID: job.Name,
		}
	}

	t.RespondCreated(c, resp)
}

// Calls the UpdateAllFactrakSurveyDeficits from Students
// TODO: This endpoint runs very slow; may want to return a job ID and then be able to query the job log (Aidan: like I did for brkt)
// TODO: If we do have a job setup, will need to make it work with multiple deployments/scaling
// UpdateAllFactrakSurveyDeficits godoc
// @Summary Updates all factrak survey deficits
// @Description runs a slow job that updates all student factrak survey deficits
// @ID update-all-factrak-survey-deficits
// @Tags admin
// @Accept  json
// @Produce  json
// @Success 200 "Job completed"
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /admin/update-all-factrak-survey-deficits [post]
func (t *Controller) UpdateAllFactrakSurveyDeficits(c *gin.Context) {
	err := t.studentModel.UpdateAllFactrakSurveyDeficits()

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
