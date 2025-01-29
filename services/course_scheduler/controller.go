package course_scheduler

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type CourseSchedulerController struct {
	services.BaseController

	courseSchedulerSelectionModel *models.CourseSchedulerSelectionModel
	userModel                     *models.UserModel
	courseModel                   *models.CourseModel
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *CourseSchedulerController {
	return &CourseSchedulerController{
		BaseController:                services.BaseController{Log: log},
		courseSchedulerSelectionModel: models.NewCourseSchedulerSelectionModel(db, log),
		userModel:                     models.NewUserModel(db, log),
		courseModel:                   models.NewCourseModel(db, log),
	}
}

// List courseSchedulerSelections
// ListCourseSchedulerSelections godoc
// @Summary List courseSchedulerSelections
// @Description lists courseSchedulerSelections (user-course-time pairs)
// @ID courseSchedulerSelections-get
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param userID query uint false "UserID"
// @Param semester query string false "Semester" Enums(FALL,WINTER,SPRING)
// @Param year query uint false "Year"
// @Success 200 {array} models.CourseSchedulerSelection
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-scheduler/selections [get]
func (t *CourseSchedulerController) ListCourseSchedulerSelections(c *gin.Context) {
	var courseSchedulerSelections []*models.CourseSchedulerSelection
	var err error

	opts := models.GetAllCourseSchedulerSelectionsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if !auth.HasScope(c, auth.ScopeCourseSchedulerAdmin) {
		userID := services.GetUserID(c)
		opts.UserID = &userID
	}

	err = t.courseSchedulerSelectionModel.GetAllCourseSchedulerSelections(&courseSchedulerSelections, &opts)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, courseSchedulerSelections)

}

type CourseSchedulerSelectionCreateParams struct {
	Semester models.SemesterType `json:"semester" binding:"required" enums:",FALL,WINTER,SPRING"`
	Year     uint                `json:"year" binding:"required"`
	CourseID uint                `json:"courseID" binding:"required"`
	Hidden   bool                `json:"hidden"`
	ProfessorID uint 			 `json:"professorID" binding:"required"`
	PeoplesoftID uint 			 `json:"peoplesoftID" binding:"required"`
	DepartmentPrefix string 	 `json:"department" binding:"required"`
}

// Add one courseSchedulerSelection entry
// CreateCourseSchedulerSelection godoc
// @Summary Add courseSchedulerSelection
// @Description add one courseSchedulerSelection entry (user-course-time pair)
// @ID courseSchedulerSelections-post
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param userID query uint false "UserID"
// @Param createParams body CourseSchedulerSelectionCreateParams true "Course scheduler selection object"
// @Success 201 {object} services.BaseResponse
// @Failure 2231 {object} services.BaseErrorResponse "course id not found"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 403 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-scheduler/selections [post]
func (t *CourseSchedulerController) CreateCourseSchedulerSelection(c *gin.Context) {
	var err error
	var params CourseSchedulerSelectionCreateParams
	if err = c.ShouldBindJSON(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var userID uint
	if auth.HasScope(c, auth.ScopeCourseSchedulerAdmin) {
		opts := models.GetAllCourseSchedulerSelectionsOptions{}
		if err = c.ShouldBindQuery(&opts); err != nil {
			t.RespondBadBind(c, err)
			return
		}
		userID = *opts.UserID
	} else {
		userID = services.GetUserID(c)
	}

	if exist, err := t.courseModel.DoesCourseExist(params.CourseID); !exist || err != nil {
		t.RespondAPIError(c, lib.ErrorCourseSchedulerInvalidCourseID)
		return
	}

	err = t.courseSchedulerSelectionModel.CreateSelection(models.CourseSchedulerSelection{
		UserID:   &userID,
		CourseID: &params.CourseID,
		Hidden:   params.Hidden,
		ProfessorID: &params.ProfessorID,
		PeoplesoftID: &params.PeoplesoftID,
		DepartmentPrefix: &params.DepartmentPrefix,
		Semester: params.Semester,
		Year:     &params.Year,
	})
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, nil)
}

// Remove courseSchedulerSelection entries using filters and delete by ID
// DeleteCourseSchedulerSelections godoc
// @Summary Delete courseSchedulerSelection entries using filters
// @Description Delete courseSchedulerSelection entries using filters
// @ID courseSchedulerSelections-delete
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param selectionID path uint true "Selection ID"
// @Success 200 {object} services.BaseResponse
// @Failure 2230 {object} services.BaseErrorResponse "attempted to access selection of unauthorized user"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-scheduler/selections/{selectionID} [delete]
func (t *CourseSchedulerController) DeleteCourseSchedulerSelections(c *gin.Context) {
	selectionID, err := services.GetUIntParam(c, "selectionID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// if user is not admin, check that they are deleting their own selection
	if !auth.HasScope(c, auth.ScopeCourseSchedulerAdmin) {
		if hasUserID, err := t.courseSchedulerSelectionModel.VerifySelectionHasUserID(services.GetUserID(c), selectionID); !hasUserID || err != nil {
			t.RespondAPIError(c, lib.ErrorCourseSchedulerUnauthorizedUser)
			return
		}
	}

	err = t.courseSchedulerSelectionModel.DeleteSelectionByID(selectionID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)

}

type CourseSchedulerSelectionUpdateParams struct {
	Hidden *bool `json:"hidden" binding:"required"`
}

// Update a course scheduler selection's hidden status
// UpdateHiddenCourseSchedulerSelection godoc
// @Summary Modify courseSchedulerSelection hidden
// @Description updates a course scheduler selection's hidden status
// @ID courseSchedulerSelections-update
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param selectionID path uint true "Selection ID"
// @Param updateParams body CourseSchedulerSelectionUpdateParams true "Request body"
// @Success 200 {object} services.BaseResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-scheduler/selections/{selectionID} [patch]
func (t *CourseSchedulerController) UpdateHiddenCourseSchedulerSelection(c *gin.Context) {
	var err error
	var params CourseSchedulerSelectionUpdateParams
	if err = c.ShouldBindJSON(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	selectionID, err := services.GetUIntParam(c, "selectionID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// if user is not admin, check that they are deleting their own selection
	if !auth.HasScope(c, auth.ScopeCourseSchedulerAdmin) {
		if hasUserID, err := t.courseSchedulerSelectionModel.VerifySelectionHasUserID(services.GetUserID(c), selectionID); !hasUserID || err != nil {
			t.RespondAPIError(c, lib.ErrorCourseSchedulerUnauthorizedUser)
			return
		}
	}

	err = t.courseSchedulerSelectionModel.SetSelectionHiddenByID(uint(selectionID), *params.Hidden)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
