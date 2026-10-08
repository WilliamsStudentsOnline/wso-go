package api

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	cfg       *config.Config
	DB        *gorm.DB
	userModel *models.UserModel
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		DB:             db,
		cfg:            cfg,
		userModel:      models.NewUserModel(db, log),
	}
}

// Checks if the passed identity token is valid and if user should be authenticated.
func (t *Controller) Authenticator(c *gin.Context) (interface{}, error) {
	// The payload of data we return
	payload := &auth.AuthenticatorPayload{
		TokenLevel: auth.TokenLevel(c.GetInt("tokenLevel")),
	}

	userID := services.GetUserID(c)

	if payload.TokenLevel == auth.TokenLevelUser {
		payload.User = &models.User{}
		err := t.DB.First(payload.User, userID).Error
		if err != nil {
			return nil, err
		}
	}

	return payload, nil
}

// AuthAPIToken godoc
// @Summary Authenticate API Token
// @Description issues an API JWT given an identity token.
// @ID getAPIToken
// @Tags auth, auth-2.0
// @Accept  json
// @Produce  json
// @Param identityToken query string true "Identity Token"
// @Success 200 {object} auth.AuthResponse
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 401 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Router /auth/api/token [post]
func authAPIToken() {}

// AuthAPIRefresh godoc
// @Summary Refresh API Token
// @description issues an API JWT by taking an existing API JWT and updating the fields. This calls the db, so it will actually modify the token's payload.
// @ID updateAPIToken
// @Tags auth, auth-2.0
// @Accept  json
// @Produce  json
// @Success 200 {object} auth.AuthResponse
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 401 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /auth/api/refresh [get]
func authAPIRefresh() {}
