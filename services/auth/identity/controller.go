package identity

import (
	"errors"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Payload struct {
	TokenLevel auth.TokenLevel
	UserID     uint
}

type Controller struct {
	services.BaseController
	cfg       *config.Config
	DB        *gorm.DB
	userModel *models.UserModel
}

type Credentials struct {
	UnixID   string `form:"unixID" json:"unixID"`
	Password string `form:"password" json:"password"`
	// If true, will authenticate based on IP. Will return either off-campus or on-campus token
	UseIP bool `form:"useIP" json:"useIP"`
	// If true, will authenticate based on IP. Fail if cannot get on-campus token.
	IsLocalIP bool `form:"localIP" json:"localIP"`
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

// Checks if passed login credentials are valid and if user should be authenticated.
func (t *Controller) Authenticator(c *gin.Context) (interface{}, error) {
	// Bind the POST parameters
	var loginVals Credentials
	if err := c.ShouldBind(&loginVals); err != nil {
		return nil, lib.ErrorMissingIdentityCredentials
	}

	// The payload of data we return
	payload := &Payload{
		TokenLevel: auth.TokenLevelUnauthenticated,
	}

	// If client is requesting a off-campus/on-campus JWT (aka client is on/off campus and wants read-only WSO access)
	if loginVals.UseIP || loginVals.IsLocalIP {
		if auth.OnCampusIP(c.ClientIP()) {
			payload.TokenLevel = auth.TokenLevelOnCampus
			return payload, nil
		} else if loginVals.IsLocalIP {
			return nil, errors.New("could not verify on-campus IP")
			// If we require it to be a local network token, error here
		}

		// Otherwise, instead of an error, sign a token for off-campus IP
		payload.TokenLevel = auth.TokenLevelOffCampus
		return payload, nil
	}

	if loginVals.UnixID == "" || loginVals.Password == "" {
		return nil, lib.ErrorMissingIdentityCredentials
	}

	// The user model interface to return if we can authenticate
	user := new(models.User)

	// Otherwise, assume client wants to authenticate with unix and password
	unixID := loginVals.UnixID
	password := loginVals.Password

	// If LDAP has been disabled (and just to be sure, environment is not production),
	// authenticate by seeing if the user exists in the database.
	if t.cfg.DisableLDAP && !t.cfg.IsProduction() {
		err := t.DB.Where(&models.User{UnixID: unixID}).First(user).Error
		if err != nil {
			return nil, lib.ErrorFailedIdentityAuthentication
		}
		payload.TokenLevel = auth.TokenLevelUser
		payload.UserID = user.ID
		return payload, nil
	}

	// Assuming we are not doing an internal network authentication, and LDAP is not disabled, do LDAP authentication
	isAuthed, err := auth.OITAuth(unixID, password)
	if err != nil {
		// Record the error in the log, as it is an internal server error (but response will be an unauthorized error)
		_ = c.Error(err)
		return nil, err
	}
	// If client is not authenticated,
	if !isAuthed {
		return nil, lib.ErrorFailedIdentityAuthentication
	}

	// Check if user exists in db and create the entry if it doesnt exist in db
	user, err = t.userModel.FirstOrCreateFromUnixID(unixID, t.cfg)
	if err != nil {
		// Record the error in the log, as it is an internal server error (but response will be an unauthorized error)
		_ = c.Error(err)
		return nil, err
	}

	payload.TokenLevel = auth.TokenLevelUser
	payload.UserID = user.ID

	return payload, nil
}

// AuthIdentityToken godoc
// @Summary Authenticate Identity Token
// @Description issues an Identity JWT given identity credentials.
// @ID auth-identity-token
// @Tags auth, auth-2.0
// @Accept  json
// @Produce  json
// @Param identityCredentials body identity.Credentials true "Identity Credentials"
// @Success 200 {object} auth.AuthResponse
// @Failure 400 {object} lib.APIError
// @Failure 401 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Router /auth/identity/token [post]
func authIdentityToken() {}
