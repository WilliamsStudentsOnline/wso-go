package auth

import (
	"errors"
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var ErrorFailedAuthentication = errors.New("incorrect unix id or password")
var ErrorMissingLoginValues = errors.New("missing unix id or password")

// Parameters passed from client when logging in
type Login struct {
	UnixID   string `form:"unix_id" json:"unix_id" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Local    bool   `form:"local" json:"local"`
}

type Controller struct {
	services.BaseController
	cfg *config.Config
	DB *gorm.DB
}

// Construct a new user controller
func NewController(cfg *config.Config, db *gorm.DB) *Controller {
	return &Controller{
		cfg: cfg,
		DB: db,
	}
}

// Checks if passed login credentials are valid and if user should be authenticated.
func (t *Controller) Authenticator(c *gin.Context) (interface{}, error) {
	// Bind the POST parameters
	var loginVals Login
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", ErrorMissingLoginValues
	}

	// The user model interface to return if we can authenticate
	var user models.User

	// If client is requesting a local-network JWT (aka client is on campus and wants read-only WSO access)
	if loginVals.Local {
		if lib.OnCampusIP(c.ClientIP()) {
			user = models.NewUserWithID(0)
			return &user, nil
		} else {
			return nil, errors.New("could not verify on-campus IP")
		}
	}

	// Otherwise, assume client wants to authenticate with unix and password
	unixID := loginVals.UnixID
	password := loginVals.Password

	// If LDAP has been disabled (and just to be sure, environment is not production),
	// authenticate by seeing if the user exists in the database.
	if t.cfg.DisableLDAP && !t.cfg.IsProduction() {
		err := t.DB.Where(&models.User{UnixID: unixID}).First(&user).Error
		if err != nil {
			return nil, ErrorFailedAuthentication
		}
		return &user, nil
	}

	// Assuming we are not doing an internal network authentication, and LDAP is not disabled, do LDAP authentication
	isAuthed, err := AuthOIT(unixID, password)
	if err != nil {
		// Record the error in the log, as it is an internal server error (but response will be an unauthorized error)
		_ = c.Error(err)
		return nil, err
	}
	// If client is not authenticated,
	if !isAuthed {
		return nil, ErrorFailedAuthentication
	}

	// Currently, we just check if the user exists in our DB, no LDAP yet
	// In real version, do FirstOrCreate.
	err = t.DB.Where(&models.User{UnixID: unixID}).First(&user).Error
	if err != nil {
		return nil, ErrorFailedAuthentication
	}

	return &user, nil
}