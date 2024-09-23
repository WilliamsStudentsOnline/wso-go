package sanitize

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

func BookListing(listing *models.BookListing, ctx *gin.Context) {
	if listing == nil || listing.User == nil {
		return
	}
	User(listing.User, ctx)
}

func BookListings(listings []*models.BookListing, ctx *gin.Context) {
	if listings == nil {
		return
	}

	isAdmin := auth.HasScope(ctx, auth.ScopeAdminAll)
	if isAdmin {
		return
	}

	for _, listing := range listings {
		BookListing(listing, ctx)
	}
}
