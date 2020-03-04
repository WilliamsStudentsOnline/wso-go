package auth

import (
	"time"

	jwt "github.com/WilliamsStudentsOnline/gin-jwt/v2"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

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
	if v.TokenLevel >= TokenLevelUser {
		scope = append(scope, auth.ScopeAllOther)
	}

	// If user exists that we signed in with
	if v.TokenLevel >= TokenLevelUser && v.User != nil {
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
	//return time.Now().Month() == time.January || (time.Now().Month() == time.February && time.Now().Day() <= 4)
	return time.Now().Month() == time.January
}
