package coursescheduler

import (
	"strconv"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CourseSchedulerController struct {
	services.BaseController
}

type SelectionGetRequest struct {
	UserID uint `json:"userID" form:"userID" binding:"required"`
}

type SelectionSetRequest struct {
	Courses string `json:"courses" form:"courses" binding:"required"`
}

type CourseSelectionsString struct {
	Courses string `json:"courses" form:"courses"`
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *CourseSchedulerController {
	return &CourseSchedulerController{
		BaseController: services.BaseController{Log: log},
	}
}

// Get courses selected by a user
// GetCourseSelectionsByUser godoc
// @Summary Get user course selections
// @Description get courses selected by a user as a string of area of study and course ID
// @ID courseSchedulerSelections-persist-get
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param userID query uint false "UserID"
// @Success 200 {object} services.CourseSelectionsString
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-scheduler/selections [get]
func (t *CourseSchedulerController) GetCourseSelectionsByUser(c *gin.Context) {
	var err error

	opts := SelectionGetRequest{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if !auth.HasScope(c, auth.ScopeAdminAll) {
		userID := services.GetUserID(c)
		if opts.UserID != userID {
			t.RespondAPIError(c, lib.ErrorMustBeSelf)
			return
		}
	}

	userIDStr := strconv.FormatUint(uint64(opts.UserID), 10)

	client := redis_util.GetClient()
	val, err := client.Get(c, userIDStr).Result()

	if err == redis.Nil {
		t.RespondAPIError(c, lib.ErrorUserHasNoSelection)
		return
	} else if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, CourseSelectionsString{Courses: val})

}

// Set user selected courses
// SetCourseSelectionsByUser godoc
// @Summary Set user course selections
// @Description set courses selected by a user as a string of area of study and course ID
// @ID courseSchedulerSelections-persist-set
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param userID query uint false "UserID"
// @Param courses body string true "Request body"
// @Success 200
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-scheduler/selections [post]
func (t *CourseSchedulerController) SetCourseSelectionsByUser(c *gin.Context) {
	var err error

	opts := SelectionGetRequest{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if !auth.HasScope(c, auth.ScopeAdminAll) {
		userID := services.GetUserID(c)
		if opts.UserID != userID {
			t.RespondAPIError(c, lib.ErrorMustBeSelf)
			return
		}
	}

	userIDStr := strconv.FormatUint(uint64(opts.UserID), 10)

	body := SelectionSetRequest{}
	if err = c.ShouldBindJSON(&body); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	client := redis_util.GetClient()
	err = client.Set(c, userIDStr, body.Courses, 0).Err()

	if err == redis.Nil {
		t.RespondOK(c, nil)
		return
	} else if err != nil {
		t.RespondError(c, err)
		return
	}

}
