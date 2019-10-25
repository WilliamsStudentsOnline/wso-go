package sanitize

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

func User(user *models.User, ctx *gin.Context) {
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
}

func Users(users []*models.User, ctx *gin.Context) {
	isAdmin := auth.HasScope(ctx, auth.ScopeAdminAll)
	if isAdmin {
		return
	}

	for _, user := range users {
		User(user, ctx)
	}
}
