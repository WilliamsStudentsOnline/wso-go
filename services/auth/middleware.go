package auth

import (
	"errors"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type TokenLevel int

// Token signing levels for payload to give scopes
const (
	TokenLevelUnauthenticated TokenLevel = iota
	TokenLevelOffCampus
	TokenLevelOnCampus
	TokenLevelSignedIn
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
			if v, ok := data.(*AuthenticatorPayload); ok {
				var scope []string

				// By default, can access bulletins
				if v.TokenLevel >= TokenLevelOffCampus {
					scope = append(scope, auth.ScopeBulletins)
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
						if v.User.Student().Senior() || v.User.EphcatchEligibility {
							scope = append(scope, auth.ScopeEphcatch)
						}

						// For factrak, user must be student and user accepted factrak policy
						if v.User.HasAcceptedFactrakPolicy {
							// TODO: ensure limited cannot get access via preloading
							// If no factrak survey deficit, give full access
							if v.User.FactrakSurveyDeficit != nil && *v.User.FactrakSurveyDeficit == 0 {
								scope = append(scope, auth.ScopeFactrakFull)
							} else {
								// Otherwise, give limited access
								scope = append(scope, auth.ScopeFactrakLimited)
							}
						}
					}

					if v.User.HasAcceptedDormtrakPolicy {
						scope = append(scope, auth.ScopeDormtrak)
					}

					// Add admin scope
					if v.User.Admin {
						scope = append(scope, auth.ScopeAdminAll)
						scope = append(scope, auth.ScopeFactrakAdmin)
					} else if v.User.FactrakAdmin {
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

			// Set "id" -> userID as the identity in the context
			userID := claims["id"].(float64)
			return userID
		},
		// Called on login to authenticate
		Authenticator: NewController(cfg, db).Authenticator,
		// What to do when a JWT is unauthorized
		Unauthorized: func(c *gin.Context, statusCode int, errorMsg string) {
			services.Base.RespondErrorCode(c, statusCode, errors.New(errorMsg))
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

/*func RegenerateToken(mw *jwt.GinJWTMiddleware) func (c *gin.Context) (string, time.Time, error) {
	// RefreshToken refresh token and check if token is expired
	return func (c *gin.Context) (string, time.Time, error) {
		claims, err := mw.CheckIfTokenExpire(c)
		if err != nil {
			return "", time.Now(), err
		}

		// Create the token
		newToken := jwtTokenizer.New(jwtTokenizer.GetSigningMethod(mw.SigningAlgorithm))
		newClaims := newToken.Claims.(jwtTokenizer.MapClaims)

		for key := range claims {
			newClaims[key] = claims[key]
		}

		expire := mw.TimeFunc().Add(mw.Timeout)
		newClaims["exp"] = expire.Unix()
		newClaims["orig_iat"] = mw.TimeFunc().Unix()
		tokenString, err := mw.signedString(newToken)

		if err != nil {
			return "", time.Now(), err
		}

		// set cookie
		if mw.SendCookie {
			maxage := int(expire.Unix() - time.Now().Unix())
			c.SetCookie(
				mw.CookieName,
				tokenString,
				maxage,
				"/",
				mw.CookieDomain,
				mw.SecureCookie,
				mw.CookieHTTPOnly,
			)
		}

		return tokenString, expire, nil
	}
}

// Copied from github.com/appleboy/gin-jwt/v2
func signedString(mw *jwt.GinJWTMiddleware, token *jwtTokenizer.Token) (string, error) {
	var tokenString string
	var err error
	if mw.usingPublicKeyAlgo() {
		tokenString, err = token.SignedString(mw.privKey)
	} else {
		tokenString, err = token.SignedString(mw.Key)
	}
	return tokenString, err
}*/
