package auth

import (
	"time"

	jwt "github.com/WilliamsStudentsOnline/gin-jwt/v2"
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type TokenType string

var (
	TokenTypeOld      TokenType = "old"
	TokenTypeIdentity TokenType = "identity"
	TokenTypeAPI      TokenType = "api"
)

func GenerateClaimsFactory(cfg *config.Config, db *gorm.DB, log *zap.SugaredLogger) func(v *AuthenticatorPayload, tokenType TokenType) jwt.MapClaims {
	return func(v *AuthenticatorPayload, tokenType TokenType) jwt.MapClaims {
		var scope []string

		// By default, can access bulletins
		if v.TokenLevel >= TokenLevelOffCampus {
			scope = append(scope, auth.ScopeBulletin)
		}

		// If on-campus, can access user info
		// TODO: THIS WAS REMOVED TO USER TO ALLOW DATA PRIVLEDGING REQUIREMENTS TO PASS WITH EDUROAM BEING ENABLED. FIGURE THIS OUT LATER
		if v.TokenLevel >= TokenLevelUser /*TokenLevelOnCampus*/ {
			scope = append(scope, auth.ScopeUsers)
		}

		// If signed in, can access: all other
		if v.TokenLevel >= TokenLevelUser {
			scope = append(scope, auth.ScopeAllOther)
		}

		// If user exists that we signed in with
		if v.TokenLevel >= TokenLevelUser && v.User != nil {
			// Allow writing
			scope = append(scope, auth.ScopeWriteSelf, auth.ScopeChat, auth.ScopeGoodrich, auth.ScopeBulletinWrite)

			// For ephcatch and factrak, user must be a student
			if v.User.IsStudent() {
				// If user is a senior or ephcatch eligible, add ephcatch scope
				if v.User.Student().Senior() || (v.User.EphcatchEligibility != nil && *v.User.EphcatchEligibility) {
					// Ensure that it is senior week and that the user has not opted out of ephcatch
					if isSeniorWeek() && !(v.User.OptOutEphcatch != nil && *v.User.OptOutEphcatch) {
						scope = append(scope, auth.ScopeEphcatch)
					}
				}

				// Ephmatch
				// Give access to edit self profile on ephmatch to all eligible users
				scope = append(scope, auth.ScopeEphmatch)
				// If user has signed up, give access to matches
				if hasEphmatchProfile(db, v.User.ID) {
					scope = append(scope, auth.ScopeEphmatchMatches)
					// If ephmatch is open, give access to like/unlike, other profiles
					if enableEphmatch(cfg) {
						// If ephmatch is senior only, only grant scope to seniors
						// Otherwise, grant scope to everyone
						if ephmatchSeniorOnly(cfg) {
							if v.User.Student().SeniorPlus() {
								scope = append(scope, auth.ScopeEphmatchProfiles)
							}
						} else {
							scope = append(scope, auth.ScopeEphmatchProfiles)
						}
					}
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

			// Add goodrich manager scope
			for _, gmUnix := range cfg.GoodrichManagerUnixes {
				if v.User.UnixID == gmUnix {
					scope = append(scope, auth.ScopeGoodrichManager)
				}
			}

			// Look up to see if user is banned
			banInfo := models.BannedUser{}
			if isBanned(db, log, v.User.ID, &banInfo) {
				// Remove any scopes that user may be banned from
				removeBannedScope(&scope, &banInfo)
				log.Info("removing scopes from banned user: ", v.User.ID)
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
			"type":       tokenType,
		}
	}
}

// isBanned searches DB if a user is banned and returns true if banned. A pointer can be passed that will
// populate with the banning info if user is banned. On error, will log but not propogate up; assume not banned.
func isBanned(db *gorm.DB, log *zap.SugaredLogger, userID uint, banInfo *models.BannedUser) bool {
	bum := models.NewBannedUserModel(db, log)
	missing, err := bum.GetBannedUserByID(userID, banInfo)
	if err != nil {
		log.Error("error when getting banned user info", err)
		return false
	}
	return !missing
}

func removeBannedScope(scope *[]string, banInfo *models.BannedUser) {
	var newScope []string
	for _, s := range *scope {
		switch s {
		case auth.ScopeFactrakLimited, auth.ScopeFactrakFull, auth.ScopeFactrakAdmin:
			if banInfo.Factrak {
				continue
			}
		case auth.ScopeDormtrak, auth.ScopeDormtrakWrite:
			if banInfo.Dormtrak {
				continue
			}
		case auth.ScopeEphcatch:
			if banInfo.Ephcatch {
				continue
			}
		case auth.ScopeBulletin:
			if banInfo.BulletinRead {
				continue
			}
		case auth.ScopeBulletinWrite:
			if banInfo.BulletinWrite {
				continue
			}
		case auth.ScopeEphmatch, auth.ScopeEphmatchMatches, auth.ScopeEphmatchProfiles:
			if banInfo.Ephmatch {
				continue
			}
		}

		newScope = append(newScope, s)
	}

	// Reassign the scope pointer to the new scope
	*scope = newScope
}

func hasEphmatchProfile(db *gorm.DB, userID uint) bool {
	var count int
	err := db.Model(&models.EphmatchProfile{}).Where("ephmatch_profiles.user_id = ?", userID).Count(&count).Error
	if err != nil {
		return false
	}

	return count > 0
}

func isSeniorWeek() bool {
	now := time.Now()
	seniorWeek := time.Date(now.Year(), time.May, 15, 0, 0, 0, 0, now.Location())
	seniorWeekEnd := time.Date(now.Year(), models.StudentCutoffMonth, 1, 0, 0, 0, 0, now.Location())
	return now.After(seniorWeek) && now.Before(seniorWeekEnd)
}

func getCurrentEphmatchEra(cfg *config.Config) (enabled bool, era *config.EphmatchEra) {
	if cfg.EphmatchEnableNow {
		return true, nil
	}

	now := time.Now()
	for _, era := range cfg.EphmatchEras {
		if era.Start.Before(now) && era.End.After(now) {
			return true, &era
		}
	}

	return false, nil
}

func enableEphmatch(cfg *config.Config) bool {
	enabled, _ := getCurrentEphmatchEra(cfg)
	return enabled
}

func ephmatchSeniorOnly(cfg *config.Config) bool {
	_, era := getCurrentEphmatchEra(cfg)
	return era != nil && era.SeniorOnly
}

func isWinterStudy() bool {
	//return time.Now().Month() == time.January || (time.Now().Month() == time.February && time.Now().Day() <= 4)
	return time.Now().Month() == time.January
}
