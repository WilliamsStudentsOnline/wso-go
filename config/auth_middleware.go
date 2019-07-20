package config

import (
	"errors"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/controllers"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/WilliamsStudentsOnline/wso-go/models"
)

type Login struct {
	UnixID   string `form:"unix_id" json:"unix_id" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Local bool `form:"local" json:"local"`
}

func LoadAuthMiddleware(cfg *Config, db *gorm.DB) (authMiddleware *jwt.GinJWTMiddleware, err error) {
	// The JWT middleware
	authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       cfg.JWTRealm,
		Key:         []byte(cfg.JWTSecretKey),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: "id",
		// Called on login to create JWT payload
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			// We take the data (which is a User) and create the payload
			if v, ok := data.(*models.User); ok {
				// Set scopes here
				scope := []string{ScopeReadAll}

				if v.ID > 0 {
					scope = append(scope, ScopeWriteSelf)
				}
				if v.Admin {
					scope = append(scope, ScopeAdminAll)
					scope = append(scope, ScopeAdminFactrak)
				}
				if v.FactrakAdmin {
					scope = append(scope, ScopeAdminFactrak)
				}

				// This is the final payload
				return jwt.MapClaims{
					"id":    v.ID,
					"scope": scope,
				}
			}
			return jwt.MapClaims{}
		},
		// Called every request to get user's id
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			user := new(models.User)
			user.ID = uint(claims["id"].(float64))
			return user
		},
		// Called on login to authenticate
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals Login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			if loginVals.Local {
				if lib.OnCampusIP(c.ClientIP()) {
					user := models.NewUserWithID(0)
					return &user, nil
				} else {
					return nil, errors.New("could not verify on-campus IP")
				}
			}

			unixID := loginVals.UnixID
			password := loginVals.Password

			// Do LDAP Authentication HERE
			_ = password

			// Currently, we just check if the user exists in our DB, no LDAP yet
			var user models.User
			// In real version, do FirstOrCreate
			err := db.Where(&models.User{UnixID: unixID}).First(&user).Error
			if err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			return &user, nil
		},
		// What to do when a JWT is unauthorized
		Unauthorized: func(c *gin.Context, statusCode int, errorMsg string) {
			controllers.Base.RespondError(statusCode, errors.New(errorMsg), c)
		},
		// Called every request; ignore this for now
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	return
}
