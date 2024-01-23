package course_scheduler

import (
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
// @ID courseSchedulerSelections-list
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param userID query uint false "Student UUID"
// @Param semester query string false "Semester" Enums(FALL,WINTER,SPRING)
// @Param year query uint false "Year"
// @Success 200 {array} models.CourseSchedulerSelection
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
func (t *CourseSchedulerController) ListCourseSchedulerSelections(c *gin.Context) {
	var courseSchedulerSelections []*models.CourseSchedulerSelection
	var err error

	opts := models.GetAllCourseSchedulerSelectionsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if opts.UserID != nil && opts.Semester != models.SemesterUndefined && opts.Year != nil {
		err = t.courseSchedulerSelectionModel.GetSelectionsByUserIDAndSemesterAndYear(*opts.UserID, opts.Semester, *opts.Year, &courseSchedulerSelections)
	} else if opts.UserID != nil {
		err = t.courseSchedulerSelectionModel.GetSelectionsByUserID(*opts.UserID, &courseSchedulerSelections)
	} else {
		err = t.courseSchedulerSelectionModel.GetAllCourseSchedulerSelections(&courseSchedulerSelections, &opts)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, courseSchedulerSelections)

}

// Add one courseSchedulerSelection entry
// AddCourseSchedulerSelection godoc
// @Summary Add courseSchedulerSelection
// @Description add one courseSchedulerSelection entry (user-course-time pair)
// @ID courseSchedulerSelections-add
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param userID query uint false "Student UUID"
// @Param courseID query uint false "Course UUID"
// @Param hidden query bool false "Hidden"
// @Param semester query string false "Semester" Enums(FALL,WINTER,SPRING)
// @Param year query uint false "Year"
// @Success 200 {object} services.BaseResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
func (t *CourseSchedulerController) AddCourseSchedulerSelection(c *gin.Context) {
	var err error
	opts := models.GetAllCourseSchedulerSelectionsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if opts.UserID != nil && opts.CourseID != nil {
		var user models.User
		usererr := t.userModel.GetUserByID(*opts.UserID, &user)
		if usererr != nil {
			t.RespondAPIError(c, lib.ErrorCourseSchedulerInvalidUserID)
		}

		var course models.Course
		courseerr := t.courseModel.GetCourseByID(*opts.CourseID, &course)
		if courseerr != nil {
			t.RespondAPIError(c, lib.ErrorCourseSchedulerInvalidCourseID)
			return
		}

		if opts.Semester != models.SemesterUndefined && opts.Year != nil {
			err = t.courseSchedulerSelectionModel.CreateSelection(models.CourseSchedulerSelection{
				User:     &user,
				UserID:   opts.UserID,
				Course:   &course,
				CourseID: opts.CourseID,
				Hidden:   opts.Hidden,
				Semester: opts.Semester,
				Year:     opts.Year,
			})
		} else {
			t.RespondAPIError(c, lib.ErrorCourseSchedulerMissingSemesterYear)
			return
		}
	} else {
		if opts.UserID == nil {
			t.RespondAPIError(c, lib.ErrorCourseSchedulerMissingUserID)
		}
		if opts.CourseID == nil {
			t.RespondAPIError(c, lib.ErrorCourseSchedulerMissingCourseID)
		}
		return
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, services.BaseResponse{
		Status: 200,
	})
}

// Remove courseSchedulerSelection entries using filters
// RemoveCourseSchedulerSelections godoc
// @Summary Remove courseSchedulerSelection entries using filters
// @Description Remove courseSchedulerSelection entries using filters
// @ID courseSchedulerSelections-remove
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param userID query uint false "Student UUID"
// @Param courseID query uint false "Course UUID"
// @Param semester query string false "Semester" Enums(FALL,WINTER,SPRING)
// @Param year query uint false "Year"
// @Success 200 {object} services.BaseResponse
// @Success 500 {object} services.BaseErrorResponse
func (t *CourseSchedulerController) RemoveCourseSchedulerSelections(c *gin.Context) {
	var err error

	opts := models.GetAllCourseSchedulerSelectionsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if opts.UserID != nil {
		if opts.CourseID != nil {
			err = t.courseSchedulerSelectionModel.DeleteAllSelectionsByUserIDAndCourseID(*opts.UserID, *opts.CourseID)
		} else if opts.Semester != models.SemesterUndefined && opts.Year != nil {
			err = t.courseSchedulerSelectionModel.DeleteAllSelectionsByUserIDAndSemesterAndYear(*opts.UserID, opts.Semester, *opts.Year)
		}
	} else if opts.CourseID != nil {
		err = t.courseSchedulerSelectionModel.DeleteAllSelectionsByCourseID(*opts.CourseID)
	} else if opts.Semester != models.SemesterUndefined && opts.Year != nil {
		err = t.courseSchedulerSelectionModel.DeleteAllSelectionsBySemesterAndYear(opts.Semester, *opts.Year)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)

}

// Update a course scheduler selection's hidden status
// HideCourseSchedulerSelection godoc
// @Summary Modify courseSchedulerSelection hidden
// @Description updates a course scheduler selection's hidden status
// @ID courseSchedulerSelections-list
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param hidden query bool false "Hidden"
// @Param userID query uint false "Student UUID"
// @Param courseID query uint false "Course UUID"
// @Param semester query string false "Semester" Enums(FALL,WINTER,SPRING)
// @Param year query uint false "Year"
// @Success 200 {object} services.BaseResponse
// @Success 500 {object} services.BaseErrorResponse
// @Security Bearer
func (t *CourseSchedulerController) HideCourseSchedulerSelection(c *gin.Context) {
	var err error

	opts := models.GetAllCourseSchedulerSelectionsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if opts.UserID != nil {
		if opts.Semester != models.SemesterUndefined && opts.Year != nil {
			err = t.courseSchedulerSelectionModel.SetSelectionHiddenByUserIDAndSemesterAndYear(*opts.UserID, opts.Semester, *opts.Year, opts.Hidden)
		} else if opts.CourseID != nil {
			err = t.courseSchedulerSelectionModel.SetSelectionHiddenByUserIDAndCourseID(*opts.UserID, *opts.CourseID, opts.Hidden)
		}
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, services.BaseResponse{
		Status: 200,
	})
}
