package sanitize

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

// Sanitize a single Ephmatch profile using the gin context.
// We choose to not sanitize if the request is self or admin.
// Otherwise, we hide sensitive information about current location
func EphmatchProfile(profile *models.EphmatchProfile, ctx *gin.Context) {
	if profile == nil {
		return
	}

	isSelf := auth.CheckIDIsSelf(ctx, profile.UserID)
	isAdmin := auth.HasScope(ctx, auth.ScopeAdminAll)

	if isSelf || isAdmin {
		return
	}

	// Current Location visibility
	if profile.LocationVisible != nil && !*profile.LocationVisible {
		profile.LocationCountry = nil
		profile.LocationState = nil
		profile.LocationTown = nil
	}

}

// Sanitize multiple Ephmatch profiles using the gin context.
// Under the hood, this just calls EphmatchProfile().
func EphmatchProfiles(profiles []*models.EphmatchProfile, ctx *gin.Context) {
	if profiles == nil {
		return
	}

	isAdmin := auth.HasScope(ctx, auth.ScopeAdminAll)
	if isAdmin {
		return
	}

	for _, profile := range profiles {
		EphmatchProfile(profile, ctx)
	}
}
