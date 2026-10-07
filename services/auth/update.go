package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// UpdateHandler can be used to update a token. The token needs a valid signature, but can be expired
// as long as it's within MaxRefresh, so put it outside the JWT middleware if expired tokens should work.
// Reply goes through the middleware's RefreshResponse.
func UpdateHandler(mw *jwt.GinJWTMiddleware, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, expire, err := UpdateToken(mw, db, c)
		if err != nil {
			c.Header("WWW-Authenticate", "JWT realm="+mw.Realm)
			if !mw.DisabledAbort {
				c.Abort()
			}
			mw.Unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, c))
			return
		}

		mw.RefreshResponse(c, http.StatusOK, tokenString, expire)
	}
}

// Basically refresh token but we can change the payload
func UpdateToken(mw *jwt.GinJWTMiddleware, db *gorm.DB, c *gin.Context) (string, time.Time, error) {
	claims, err := mw.CheckIfTokenExpire(c)
	if err != nil {
		return "", time.Now(), err
	}

	// Get the new payload
	payload, err := updateClaims(claims, db, c)
	if err != nil {
		return "", time.Now(), err
	}

	// Create the token
	return mw.TokenGenerator(payload)
}

// Update token calls this. It passes in identity data, like authorizor, and the gin
// context. From there, the function should work somewhat like payload func to generate
// a new payload.
func updateClaims(claims map[string]interface{}, db *gorm.DB, c *gin.Context) (*AuthenticatorPayload, error) {
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
	} else if tokenLevel == TokenLevelUser {
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

	return payload, nil
}
