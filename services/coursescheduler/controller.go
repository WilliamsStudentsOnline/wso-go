package coursescheduler

import (
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

var client *redis_util.RedisClient

func init() {
	ConfigureClientForProd()
}

func ConfigureClientForProd() error {
	var err error
	client, err = redis_util.SetupClient("localhost:6739", "", redis_util.CourseSchedulerSelectionsDatabaseID)
	return err
}

func ConfigureControllerForTest() error {
	var err error
	client, err = redis_util.SetupRedisClientForTest()
	return err
}

type CourseSchedulerController struct {
	services.BaseController
}

type SelectionSetRequest struct {
	Courses string `json:"courses" form:"courses" binding:"required"`
}

func (m SelectionSetRequest) MarshalBinary() ([]byte, error) {
	return []byte(m.Courses), nil
}

type CourseSelectionsString struct {
	Courses string `json:"courses"`
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
// @Param userID path uint true "UserID"
// @Success 200 {object} services.CourseSelectionsString
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-selections/{userID} [get]
func (t *CourseSchedulerController) GetCourseSelectionsByUser(c *gin.Context) {
	var err error

	userID, err := services.GetUIntParam(c, "userID")
	if err != nil {
		t.RespondAPIError(c, lib.ErrorUserIDNoParse)
		return
	}

	if !auth.HasScope(c, auth.ScopeAdminAll) {
		authedUserID := services.GetUserID(c)
		if userID != authedUserID {
			t.RespondAPIError(c, lib.ErrorMustBeSelf)
			return
		}
	}

	userIDStr := redis_util.GetUserSelectionStr(userID)
	val, err := client.Get(c, userIDStr)
	if err == redis_util.ErrRedisClientNotConfigured {
		t.RespondAPIError(c, lib.ErrorRedisClientNotConfigured)
	}
	valStr, ok := val.(string)
	if !ok {
		t.RespondAPIError(c, lib.ErrorCourseSchedulerSelectionStrconv)
		return
	}

	if err == redis.Nil {
		t.RespondOK(c, CourseSelectionsString{Courses: ""})
		return
	} else if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, CourseSelectionsString{Courses: valStr})

}

// Set user selected courses
// SetCourseSelectionsByUser godoc
// @Summary Set user course selections
// @Description set courses selected by a user as a string of area of study and course ID
// @ID courseSchedulerSelections-persist-set
// @Tags course-scheduler
// @Accept json
// @Produce json
// @Param userID path uint true "UserID"
// @Param courses body string true "Request body"
// @Success 200
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /course-selections/{userID} [put]
func (t *CourseSchedulerController) SetCourseSelectionsByUser(c *gin.Context) {
	var err error

	userID, err := services.GetUIntParam(c, "userID")
	if err != nil {
		t.RespondAPIError(c, lib.ErrorUserIDNoParse)
		return
	}

	if !auth.HasScope(c, auth.ScopeAdminAll) {
		authedUserID := services.GetUserID(c)
		if userID != authedUserID {
			t.RespondAPIError(c, lib.ErrorMustBeSelf)
			return
		}
	}

	userIDStr := redis_util.GetUserSelectionStr(userID)
	body := SelectionSetRequest{}
	if err = c.ShouldBindJSON(&body); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	err = client.Set(c, userIDStr, body)
	if err == redis_util.ErrRedisClientNotConfigured {
		t.RespondAPIError(c, lib.ErrorRedisClientNotConfigured)
	}

	if err == nil {
		t.RespondOK(c, nil)
		return
	} else {
		t.RespondError(c, err)
		return
	}

}
