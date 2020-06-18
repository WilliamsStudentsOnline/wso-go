package ahe2nht1

import (
	"errors"
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController                   // Inherit the base controller
	userModel               *models.UserModel // DB communication to get user info
}

// Construct a new onboarding controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		userModel:      models.NewUserModel(db, log),
	}
}

// GetUserByUnix godoc
// @Summary Gets a user
// @Description Gets a user from their unix. Onboarding exercise.
// @ID onboarding-ahe2nht1-get-user-by-unix
// @Tags onboarding
// @Accept  json
// @Produce  json
// @Param unixID path string true "Unix ID"
// @Success 200 {object} models.User
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /onboarding/ahe2nht1/{unixID} [get]
func (t *Controller) GetUserByUnix(c *gin.Context) {
	id := c.Param("unixID")

	if id == "" {
		t.RespondErrorCode(c, http.StatusBadRequest, errors.New("missing unixID"))
		return
	}

	var user models.User
	//assume no error for now
	err := t.userModel.GetUserByUnixID(id, &user)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	sanitize.User(&user, c)

	t.RespondOK(c, user)
}
