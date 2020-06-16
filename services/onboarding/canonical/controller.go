package canonical

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
// @ID onboarding-canonical-get-user-by-unix
// @Tags onboarding
// @Accept  json
// @Produce  json
// @Param unixID path string true "Unix ID"
// @Success 200 {object} models.User
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /onboarding/canonical/{unixID} [get]
func (t *Controller) GetUserByUnix(c *gin.Context) {
	var user models.User
	var err error

	// Get the passed unixID parameter
	unixID := c.Param("unixID")
	if unixID == "" {
		t.RespondErrorCode(c, http.StatusBadRequest, errors.New("missing unixID"))
		return
	}

	// Get the user from the database
	err = t.userModel.GetUserByUnixID(unixID, &user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Sanitize the user from any secret info
	sanitize.User(&user, c)

	// Respond to the API call with 200 OK and the data.
	t.RespondOK(c, user)
}
