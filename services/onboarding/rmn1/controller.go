package rmn1

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
// @ID onboarding-rmn1-get-user-by-unix
// @Tags onboarding
// @Accept  json
// @Produce  json
// @Param unixID path string true "Unix ID"
// @Success 200 {object} models.User
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /onboarding/rmn1/{unixID} [get]
func (t *Controller) GetUserByUnix(c *gin.Context) {
	var err error        // Error
	var user models.User // User

	userUnix := c.Param("unixID") // Get user's unixID
	if len(userUnix) == 0 {
		// Respond with error if the unixID is invalid
		err = errors.New("userUnix not received")
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Get the user by unixID
	err = t.userModel.GetUserByUnixID(userUnix, &user)

	if err != nil {
		// Respond with error if the user is not found
		t.RespondError(c, err)
		return
	}

	// Sanitize the user of any secret information
	sanitize.User(&user, c)

	// Respond OK if there is no error
	t.RespondOK(c, user)

}
