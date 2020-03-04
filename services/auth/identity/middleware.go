package identity

import (
	"errors"
	"time"

	jwt "github.com/WilliamsStudentsOnline/gin-jwt/v2"
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func LoadMiddleware(cfg *config.Config, db *gorm.DB, log *zap.SugaredLogger) (authMiddleware *jwt.GinJWTMiddleware, err error) {

	// The JWT middleware
	authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm: cfg.JWTRealm,

		// Signing algorithm setup. Contains both secret key and pub/priv keys
		SigningAlgorithm: cfg.JWTSigningAlgo,
		Key:              []byte(cfg.Secrets.JWTSecretKey),
		PubKeyFile:       cfg.JWTPublicKeyFile,
		PrivKeyFile:      cfg.JWTPrivateKeyFile,

		// Refreshing and timeout have same duration
		Timeout: time.Duration(cfg.JWTIdentityTimeoutHours) * time.Hour,
		// No refreshing
		MaxRefresh: 0,

		// How are we distinguishing signed in users: by user id
		IdentityKey: "id",

		// Called on login to create JWT payload
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			// We take the data (which is a User) and create the payload
			if v, ok := data.(*Payload); ok {
				return GenerateClaims(v)
			}
			return jwt.MapClaims{}
		},

		// Called every request to get user's id
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)

			// Put the token level as a scope just for the authorizor
			c.Set("tokenLevel", int(claims["tokenLevel"].(float64)))

			// Set "id" -> userID as the identity in the context
			userID := uint(claims["id"].(float64))
			return userID
		},

		// Called on login to authenticate
		Authenticator: NewController(db, cfg, log).Authenticator,

		// What to do when a JWT is unauthorized
		Unauthorized: func(c *gin.Context, statusCode int, errorMsg string) {
			if errorMsg == jwt.ErrExpiredToken.Error() {
				c.Set(services.UpdateTokenKey, true)
			}
			services.Base.RespondErrorCode(c, statusCode, errors.New(errorMsg))
		},

		// What to do when a login works
		LoginResponse: func(c *gin.Context, statusCode int, token string, expire time.Time) {
			services.Base.RespondOK(c, auth.AuthResponse{
				Token:  token,
				Expire: expire,
			})
		},

		// What to do when a refresh works
		RefreshResponse: func(c *gin.Context, statusCode int, token string, expire time.Time) {
			services.Base.RespondOK(c, auth.AuthResponse{
				Token:  token,
				Expire: expire,
			})
		},

		// Called every request; ignore this for now
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// Only allow token level authenticated and above
			tokenLevel := c.GetInt("tokenLevel")
			if auth.TokenLevel(tokenLevel) == auth.TokenLevelUnauthenticated {
				return false
			}
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
		TokenLookup: "header: Authorization, query: token",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	return
}

func GenerateClaims(v *Payload) jwt.MapClaims {

	// This is the final payload
	return jwt.MapClaims{
		"id":         v.UserID,
		"tokenLevel": v.TokenLevel,
	}
}
