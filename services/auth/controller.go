package auth

import (
	"errors"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var ErrorFailedAuthentication = errors.New("incorrect unix id or password")
var ErrorMissingLoginValues = errors.New("missing unix id or password")

// Parameters passed from client when logging in
type LoginParams struct {
	UnixID   string `form:"unixID" json:"unixID"`
	Password string `form:"password" json:"password"`
	// If true, will authenticate based on IP. Will return either off-campus or on-campus token
	UseIP bool `form:"useIP" json:"useIP"`
	// If true, will authenticate based on IP. Fail if cannot get on-campus token.
	IsLocalIP bool `form:"localIP" json:"localIP"`
}

type Controller struct {
	services.BaseController
	cfg       *config.Config
	DB        *gorm.DB
	userModel *models.UserModel
}

// Construct a new user controller
func NewController(cfg *config.Config, db *gorm.DB) *Controller {
	return &Controller{
		cfg:       cfg,
		DB:        db,
		userModel: models.NewUserModel(db),
	}
}

type AuthenticatorPayload struct {
	User       *models.User
	TokenLevel TokenLevel
}

// Checks if passed login credentials are valid and if user should be authenticated.
func (t *Controller) Authenticator(c *gin.Context) (interface{}, error) {
	// Bind the POST parameters
	var loginVals LoginParams
	if err := c.ShouldBind(&loginVals); err != nil {
		return nil, ErrorMissingLoginValues
	}

	// The payload of data we return
	payload := &AuthenticatorPayload{
		TokenLevel: TokenLevelUnauthenticated,
	}

	// If client is requesting a off-campus/on-campus JWT (aka client is on/off campus and wants read-only WSO access)
	if loginVals.UseIP || loginVals.IsLocalIP {
		if OnCampusIP(c.ClientIP()) {
			payload.TokenLevel = TokenLevelOnCampus
			return payload, nil
		} else if loginVals.IsLocalIP {
			return nil, errors.New("could not verify on-campus IP")
			// If we require it to be a local network token, error here
		}

		// Otherwise, instead of an error, sign a token for off-campus IP
		payload.TokenLevel = TokenLevelOffCampus
		return payload, nil
	}

	if loginVals.UnixID == "" || loginVals.Password == "" {
		return nil, ErrorMissingLoginValues
	}

	// The user model interface to return if we can authenticate
	user := new(models.User)

	// Otherwise, assume client wants to authenticate with unix and password
	unixID := loginVals.UnixID
	password := loginVals.Password

	payload.User = user

	// If LDAP has been disabled (and just to be sure, environment is not production),
	// authenticate by seeing if the user exists in the database.
	if t.cfg.DisableLDAP && !t.cfg.IsProduction() {
		err := t.DB.Where(&models.User{UnixID: unixID}).First(user).Error
		if err != nil {
			return nil, ErrorFailedAuthentication
		}
		payload.TokenLevel = TokenLevelSignedIn
		return payload, nil
	}

	// Assuming we are not doing an internal network authentication, and LDAP is not disabled, do LDAP authentication
	isAuthed, err := OITAuth(unixID, password)
	if err != nil {
		// Record the error in the log, as it is an internal server error (but response will be an unauthorized error)
		_ = c.Error(err)
		return nil, err
	}
	// If client is not authenticated,
	if !isAuthed {
		return nil, ErrorFailedAuthentication
	}

	// Check if user exists in DB and create the entry if it doesnt exist in DB
	user, err = t.userModel.FirstOrCreateFromUnixID(unixID, t.cfg)
	if err != nil {
		// Record the error in the log, as it is an internal server error (but response will be an unauthorized error)
		_ = c.Error(err)
		return nil, err
	}

	payload.TokenLevel = TokenLevelSignedIn
	payload.User = user

	return payload, nil
}

// These functions are just to make the API documentation work

// AuthLogin godoc
// @Summary Authenticate and Login
// @Description attempts to get a JWT by logging into server.
// @ID auth-login
// @Tags auth
// @Accept  json
// @Produce  json
// @Param loginParams body auth.LoginParams true "Login Parameters"
// @Success 200 {object} auth.AuthResponse
// @Failure 400 {object} lib.APIError
// @Failure 401 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Router /auth/login [post]
func authLogin() {}

// AuthUpdate godoc
// @Summary Update Token
// @description attempts to get a JWT by taking an existing JWT and updating the fields. This calls the DB, so it will actually modify the token's payload.
// @ID auth-update
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {object} auth.AuthResponse
// @Failure 400 {object} lib.APIError
// @Failure 401 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /auth/update-token [get]
func authUpdate() {}

// AuthRefresh godoc
// @Summary Refresh Token
// @Description attempts to get a JWT by taking an existing JWT and refreshing it. This will not change the payload.
// @ID auth-refresh
// @Tags auth
// @Accept  json
// @Produce  json
// @Success 200 {object} auth.AuthResponse
// @Failure 400 {object} lib.APIError
// @Failure 401 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /auth/refresh-token [get]
func authRefresh() {}
