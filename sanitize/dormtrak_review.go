package sanitize

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

// Sanitize a single dormtrak review using the gin context.
// We choose to not sanitize if the request is self or admin.
// Otherwise, we hide sensitive information about user id.
func DormtrakReview(review *models.DormtrakReview, ctx *gin.Context) {
	if review == nil {
		return
	}

	isSelf := auth.CheckIDIsSelf(ctx, review.UserID)
	isAdmin := auth.HasScope(ctx, auth.ScopeAdminAll, auth.ScopeFactrakAdmin)

	if isSelf || isAdmin {
		return
	}

	review.User = nil
	review.UserID = 0
}

// Sanitize multiple dormtrak reviews using the gin context.
// Under the hood, this just calls DormtrakReview().
func DormtrakReviews(reviews []*models.DormtrakReview, ctx *gin.Context) {
	if reviews == nil {
		return
	}

	isAdmin := auth.HasScope(ctx, auth.ScopeAdminAll, auth.ScopeFactrakAdmin)
	if isAdmin {
		return
	}

	for _, review := range reviews {
		DormtrakReview(review, ctx)
	}
}
