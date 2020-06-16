package main

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
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
// @ID onboarding-$UNIX-get-user-by-unix
// @Tags onboarding
// @Accept  json
// @Produce  json
// @Param unixID path string true "Unix ID"
// @Success 200 {object} models.User
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /onboarding/$UNIX/{unixID} [get]
func (t *Controller) GetUserByUnix(c *gin.Context) {
	id := c.Param("unixID")
	fmt.Printf("id: %+v", id)
}

func main() {
	var control Controller
	control.NewController()
	control.GetUserByUnix()
}

/*
Hints:

c.Param("unixID") should return the passed unix ID
Use userModel.GetUserByUnixID to get a user from a unix ID
Make sure to respond to errors and stop execution of the function.
Be sure to sanitize the user of any secret information with sanitize.User(&user, c)
*/
