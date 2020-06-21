package ks19

import (
	"errors"
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"net/http"
)
type Controller struct {
	services.BaseController // Inherit the base controller
	userModel *models.UserModel // DB communication to get user info
}

// Construct new controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller {
		BaseController: services.BaseController{Log: log},
		userModel: models.NewUserModel(db, log),
	}
}

// GetUserByUnix godoc
// @Summary Gets a user
// @Description Gets a user from their unix. Onboarding exercise.
// @ID onboarding-ks19-get-user-by-unix
// @Tags onboarding
// @Accept  json
// @Produce  json
// @Param unixID path string true "Unix ID"
// @Success 200 {object} models.User
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /onboarding/$UNIX/{unixID} [get]
func (t *Controller) GetUserByUnix(c *gin.Context) {
	var user models.User
	var err error

	//respond with 400 if unix is empty
	unixID := c.Param("unixID")
	if unixID == "" {
		t.RespondErrorCode(c, http.StatusBadRequest, errors.New("missing unixID") )
		return
	}
	err = t.userModel.GetUserByUnixID(unixID, &user)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	// sanitize user
	sanitize.User(&user, c)

	//respond ok
	t.RespondOK(c, user)

}