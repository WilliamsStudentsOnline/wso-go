package auth

import (
	"errors"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func LoadAuthMiddleware(cfg *config.Config, db *gorm.DB) (authMiddleware *jwt.GinJWTMiddleware, err error) {
	// The JWT middleware
	authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       cfg.JWTRealm,
		Key:         []byte(cfg.Secrets.JWTSecretKey),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: "id",
		// Called on login to create JWT payload
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			// We take the data (which is a User) and create the payload
			if v, ok := data.(*models.User); ok {
				// Set scopes here
				scope := []string{config.ScopeReadAll}

				if v.ID > 0 {
					scope = append(scope, config.ScopeWriteSelf)
				}
				if v.Admin {
					scope = append(scope, config.ScopeAdminAll)
					scope = append(scope, config.ScopeAdminFactrak)
				}
				if v.FactrakAdmin {
					scope = append(scope, config.ScopeAdminFactrak)
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
		Authenticator: NewController(cfg, db).Authenticator,
		// What to do when a JWT is unauthorized
		Unauthorized: func(c *gin.Context, statusCode int, errorMsg string) {
			services.Base.RespondError(statusCode, errors.New(errorMsg), c)
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
