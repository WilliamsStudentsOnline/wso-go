package course_scheduler

import (
	"strconv"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
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

	var userID = services.GetUserID(c)
	if opts.Semester != models.SemesterUndefined && opts.Year != nil {
		err = t.courseSchedulerSelectionModel.GetSelectionsByUserIDAndSemesterAndYear(userID, opts.Semester, *opts.Year, &courseSchedulerSelections)
	} else {
		err = t.courseSchedulerSelectionModel.GetSelectionsByUserID(userID, &courseSchedulerSelections)
	}

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
}

// Add one courseSchedulerSelection entry
// CreateCourseSchedulerSelection godoc
// @Summary Add courseSchedulerSelection
// @Description add one courseSchedulerSelection entry (user-course-time pair)
// @ID courseSchedulerSelections-post
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param requestBody body CourseSchedulerSelectionCreateParams true "Course scheduler selection object"
// @Success 200 {object} services.BaseResponse
// @Failure 2230 {object} services.BaseErrorResponse "user id not found"
// @Failure 2231 {object} services.BaseErrorResponse "course id not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-scheduler/selections [post]
func (t *CourseSchedulerController) CreateCourseSchedulerSelection(c *gin.Context) {
	var err error
	var params CourseSchedulerSelectionCreateParams
	if err = c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	userID := services.GetUserID(c)
	if exist, err := t.userModel.DoesUserExist(userID); exist || err != nil {
		t.RespondAPIError(c, lib.ErrorCourseSchedulerInvalidUserID)
		return
	}

	if exist, err := t.courseModel.DoesCourseExist(params.CourseID); exist || err != nil {
		t.RespondAPIError(c, lib.ErrorCourseSchedulerInvalidCourseID)
		return
	}

	err = t.courseSchedulerSelectionModel.CreateSelection(models.CourseSchedulerSelection{
		UserID:   &userID,
		CourseID: &params.CourseID,
		Hidden:   params.Hidden,
		Semester: params.Semester,
		Year:     &params.Year,
	})
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, services.BaseResponse{
		Status: 200,
	})
}

// Remove courseSchedulerSelection entries using filters and delete by ID
// DeleteCourseSchedulerSelections godoc
// @Summary Delete courseSchedulerSelection entries using filters
// @Description Delete courseSchedulerSelection entries using filters
// @ID courseSchedulerSelections-delete
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param id path uint true "Selection ID"
// @Success 200 {object} services.BaseResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /selections/{selectionID} [delete]
func (t *CourseSchedulerController) DeleteCourseSchedulerSelections(c *gin.Context) {
	selectionIDStr := c.Param("id")
	selectionID, err := strconv.ParseUint(selectionIDStr, 10, 64)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	err = t.courseSchedulerSelectionModel.DeleteSelectionByID(uint(selectionID))
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)

}

type CourseSchedulerSelectionUpdateParams struct {
	Hidden bool `json:"hidden" binding:"required"`
}

// Update a course scheduler selection's hidden status
// UpdateHiddenCourseSchedulerSelection godoc
// @Summary Modify courseSchedulerSelection hidden
// @Description updates a course scheduler selection's hidden status
// @ID courseSchedulerSelections-update
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param id path uint true "Selection ID"
// @Param requestBody body UpdateHiddenParams true "Request body"
// @Success 200 {object} services.BaseResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /selections/{selectionID} [patch]
func (t *CourseSchedulerController) UpdateHiddenCourseSchedulerSelection(c *gin.Context) {
	selectionIDStr := c.Param("id")
	selectionID, err := strconv.ParseUint(selectionIDStr, 10, 64)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	var params CourseSchedulerSelectionUpdateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	err = t.courseSchedulerSelectionModel.SetSelectionHiddenByID(uint(selectionID), params.Hidden)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
