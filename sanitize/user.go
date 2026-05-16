package sanitize

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

// Sanitize a single user using the gin context.
// We choose to not sanitize if the request is self or admin.
// Otherwise, we hide sensitive information about cell phone and home phones,
// and we hide dorm and home visibility if those respective flags are set.
func User(user *models.User, ctx *gin.Context) {
	if user == nil {
		return
	}

	isSelf := auth.CheckIDIsSelf(ctx, user.ID)
	isAdmin := auth.HasScope(ctx, auth.ScopeAdminAll)

	if isSelf || isAdmin {
		return
	}

	// Remove sensitive information about phones
	user.CellPhone = nil
	user.HomePhone = nil

	// Dorm visibility
	if user.DormVisible != nil && !*user.DormVisible {
		user.DormRoom = nil
		user.DormRoomID = nil
	}

	// Home visibility
	if user.HomeVisible != nil && !*user.HomeVisible {
		user.HomeCountry = nil
		user.HomeState = nil
		user.HomeTown = nil
		user.HomeZip = nil
	}

	// Dietary preference visibility
	if user.DietaryPrefVisible != nil && !*user.DietaryPrefVisible {
		user.DietaryPreference = nil
	}

	// Williams ID
	user.WilliamsID = ""
}

// Sanitize multiple users using the gin context.
// Under the hood, this just calls User().
func Users(users []*models.User, ctx *gin.Context) {
	if users == nil {
		return
	}

	isAdmin := auth.HasScope(ctx, auth.ScopeAdminAll)
	if isAdmin {
		return
	}

	for _, user := range users {
		User(user, ctx)
	}
}
