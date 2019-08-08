package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListReviews godoc
// @Summary List reviews
// @Description lists all reviews
// @ID dormtrak-list-reviews
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Param dormID query uint false "Dorm ID"
// @Param dormRoomID query uint false "Dorm Room ID"
// @Param userID query uint false "User ID"
// @Param offset query time.Time false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param commented query bool false "Restrict to commented reviews"
// @Success 200 {array} models.DormtrakReview
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /dormtrak/reviews [get]
func (t *Controller) ListReviews(c *gin.Context) {
	params := models.GetAllDormtrakReviewsOptions{}

	err := c.ShouldBindQuery(&params)
	if err != nil {
		return
	}

	// If we specify userID, ensure it is either self or we are admin
	if params.UserID != nil && *params.UserID != services.GetUserID(c) && !auth.HasScope(c, auth.ScopeAdminAll) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	var reviews []*models.DormtrakReview
	err = t.reviewModel.GetAllReviewsWithOptions(&reviews, &params)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromReviews(c, reviews)

	t.RespondOK(c, reviews)
}
