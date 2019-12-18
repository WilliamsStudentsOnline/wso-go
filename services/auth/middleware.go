package auth

import (
	"errors"
	"time"

	jwt "github.com/WilliamsStudentsOnline/gin-jwt/v2"
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type TokenLevel int

// Token signing levels for payload to give scopes
const (
	TokenLevelUnauthenticated TokenLevel = iota
	TokenLevelOffCampus
	TokenLevelOnCampus
	TokenLevelSignedIn
)

type AuthResponse struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

func LoadAuthMiddleware(cfg *config.Config, db *gorm.DB, log *zap.SugaredLogger) (authMiddleware *jwt.GinJWTMiddleware, err error) {
	algo := "HS256"
	if cfg.JWTUseAsymmetric {
		algo = "RS256"
	}

	// The JWT middleware
	authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm: cfg.JWTRealm,

		// Signing algorithm setup. Contains both secret key and pub/priv keys
		SigningAlgorithm: algo,
		Key:              []byte(cfg.Secrets.JWTSecretKey),
		PubKeyFile:       cfg.JWTPublicKeyFile,
		PrivKeyFile:      cfg.JWTPrivateKeyFile,

		// Refreshing and timeout have same duration
		Timeout:    time.Duration(cfg.JWTTimeoutHours) * time.Hour,
		MaxRefresh: time.Duration(cfg.JWTTimeoutHours) * time.Hour,

		// How are we distinguishing signed in users: by user id
		IdentityKey: "id",

		// Called on login to create JWT payload
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			// We take the data (which is a User) and create the payload
			if v, ok := data.(*AuthenticatorPayload); ok {
				return GenerateClaims(v)
			}
			return jwt.MapClaims{}
		},

		// Called every request to get user's id
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)

			// Set scope as a context variable
			jwtScopesIface := (claims["scope"]).([]interface{})

			jwtScopes := make([]string, len(jwtScopesIface))
			for i, v := range jwtScopesIface {
				jwtScopes[i] = v.(string)
			}
			c.Set("scopes", jwtScopes)

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
			services.Base.RespondOK(c, AuthResponse{
				Token:  token,
				Expire: expire,
			})
		},

		// What to do when a refresh works
		RefreshResponse: func(c *gin.Context, statusCode int, token string, expire time.Time) {
			services.Base.RespondOK(c, AuthResponse{
				Token:  token,
				Expire: expire,
			})
		},

		// Called every request; ignore this for now
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// Only allow token level authenticated and above
			tokenLevel := c.GetInt("tokenLevel")
			if TokenLevel(tokenLevel) == TokenLevelUnauthenticated {
				return false
			}
			return true
		},

		// Update token calls this. It passes in identity data, like authorizor, and the gin
		// context. From there, the function should work somewhat like payload func to generate
		// a new payload.
		UpdateClaims: func(claims jwt.MapClaims, c *gin.Context) (jwt.MapClaims, error) {
			tokenLevel := TokenLevel(claims["tokenLevel"].(float64))

			payload := new(AuthenticatorPayload)
			payload.TokenLevel = tokenLevel

			// Deal with special token-level data.
			if tokenLevel == TokenLevelOffCampus || tokenLevel == TokenLevelOnCampus {
				// If lower-level token, check if we must upgrade/downgrade the token's level
				if OnCampusIP(c.ClientIP()) {
					payload.TokenLevel = TokenLevelOnCampus
				} else {
					payload.TokenLevel = TokenLevelOffCampus
				}
			} else if tokenLevel == TokenLevelSignedIn {
				// If signed in token, get userID and find it in db.
				userID, ok := claims["id"].(float64)
				if !ok {
					return nil, errors.New("could not find user id in claim")
				}

				payload.User = &models.User{}
				err := db.First(payload.User, int(userID)).Error
				if err != nil {
					return nil, err
				}
			}

			return GenerateClaims(payload), nil
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

func GenerateClaims(v *AuthenticatorPayload) jwt.MapClaims {
	var scope []string

	// By default, can access bulletins
	if v.TokenLevel >= TokenLevelOffCampus {
		scope = append(scope, auth.ScopeBulletin)
	}

	// If on-campus, can access user info
	if v.TokenLevel >= TokenLevelOnCampus {
		scope = append(scope, auth.ScopeUsers)
	}

	// If signed in, can access: all other
	if v.TokenLevel >= TokenLevelSignedIn {
		scope = append(scope, auth.ScopeAllOther)
	}

	// If user exists that we signed in with
	if v.TokenLevel >= TokenLevelSignedIn && v.User != nil {
		// Allow writing
		scope = append(scope, auth.ScopeWriteSelf)

		// For ephcatch and factrak, user must be a student
		if v.User.IsStudent() {
			// If user is a senior or ephcatch eligible, add ephcatch scope
			if v.User.Student().Senior() || (v.User.EphcatchEligibility != nil && *v.User.EphcatchEligibility) {
				// Ensure that it is senior week and that the user has not opted out of ephcatch
				if isSeniorWeek() && !(v.User.OptOutEphcatch != nil && *v.User.OptOutEphcatch) {
					scope = append(scope, auth.ScopeEphcatch)
				}
			}

			// If month is January (winter study), add the ephmatch scope. User opts out by deleting profile. By default opt in.
			if isWinterStudy() {
				scope = append(scope, auth.ScopeEphmatch)
			}

			// For factrak, user must be student and user accepted factrak policy
			if v.User.HasAcceptedFactrakPolicy != nil && *v.User.HasAcceptedFactrakPolicy {
				// TODO: ensure limited cannot get access via preloading
				// If no factrak survey deficit, give full access
				if v.User.FactrakSurveyDeficit != nil && *v.User.FactrakSurveyDeficit == 0 {
					scope = append(scope, auth.ScopeFactrakFull)
				} else {
					// Otherwise, give limited access
					scope = append(scope, auth.ScopeFactrakLimited)
				}
			}

			// For dormtrak, user must be a student and user accepted dormtrak policy
			if v.User.HasAcceptedDormtrakPolicy != nil && *v.User.HasAcceptedDormtrakPolicy {
				scope = append(scope, auth.ScopeDormtrak)

				// If the student is upper class, they can write reviews
				if v.User.Student().IsUpperClass() {
					scope = append(scope, auth.ScopeDormtrakWrite)
				}
			}
		}

		// Add admin scope
		if v.User.Admin != nil && *v.User.Admin {
			scope = append(scope, auth.ScopeAdminAll)
			scope = append(scope, auth.ScopeFactrakAdmin)
		} else if v.User.FactrakAdmin != nil && *v.User.FactrakAdmin {
			// If not admin, check if factrak admin
			scope = append(scope, auth.ScopeFactrakAdmin)
		}
	}

	var jwtUserID uint = 0
	if v.User != nil {
		jwtUserID = v.User.ID
	}

	// This is the final payload
	return jwt.MapClaims{
		"id":         jwtUserID,
		"tokenLevel": v.TokenLevel,
		"scope":      scope,
	}
}

func isSeniorWeek() bool {
	now := time.Now()
	seniorWeek := time.Date(now.Year(), time.May, 15, 0, 0, 0, 0, now.Location())
	seniorWeekEnd := time.Date(now.Year(), models.StudentCutoffMonth, 1, 0, 0, 0, 0, now.Location())
	return now.After(seniorWeek) && now.Before(seniorWeekEnd)
}

func isWinterStudy() bool {
	return time.Now().Month() == time.January
}
