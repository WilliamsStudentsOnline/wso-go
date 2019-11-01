package bulletin

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// ListDiscussions godoc
// @Summary List discussions
// @Description lists all bulletin discussions
// @ID bulletins-list-discussions
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Success 200 {array} models.Discussion
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/discussions [get]
func (t *Controller) ListDiscussions(c *gin.Context) {
	params := models.GetAllDiscussionsOptions{}

	err := c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var discussions []*models.Discussion
	err = t.discussionModel.GetAllDiscussions(&discussions, &params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.discussionModel.CountAllDiscussions()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	// Remove user info if not a user. Need this, as bulletin service is public
	removeUserInfoFromDiscussions(c, discussions)

	t.RespondOK(c, discussions)
}

// GetDiscussion godoc
// @Summary Get discussion
// @Description Get discussion by ID with user, posts, and posts.user preloaded
// @ID bulletins-get-discussion
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param discussionID path uint true "Discussion ID"
// @Param preload query []string false "Preload List"
// @Success 200 {object} models.Discussion
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/discussions/{discussionID} [get]
func (t *Controller) GetDiscussion(c *gin.Context) {
	// Decode discussionID.
	discussionID, err := services.GetUIntParam(c, "discussionID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	params := models.GetDiscussionByIDOptions{}

	err = c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Do database query
	var discussion models.Discussion
	err = t.discussionModel.GetDiscussionByID(discussionID, &discussion, &params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Remove user info if not a user. Need this, as bulletin service is public
	if hasUserAuth(c) {
		sanitize.User(discussion.User, c)
	} else {
		discussion.User = nil
		discussion.ExUserName = ""
	}
	removeUserInfoFromPosts(c, discussion.Posts)

	t.RespondOK(c, discussion)
}

// GetDiscussionPosts godoc
// @Summary Get discussion posts
// @Description gets all posts in a discussion
// @ID bulletins-get-discussion-posts
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param discussionID path uint true "Discussion ID"
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Success 200 {array} models.Post
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/discussions/{discussionID}/posts [get]
func (t *Controller) GetDiscussionPosts(c *gin.Context) {
	// Decode discussionID.
	discussionID, err := services.GetUIntParam(c, "discussionID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Discussion must exist
	exists, err := t.discussionModel.DoesDiscussionExist(discussionID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	params := models.GetPostsByDiscussionOptions{}

	err = c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var posts []*models.Post
	err = t.postModel.GetPostsByDiscussion(discussionID, &posts, &params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Remove user info if not a user. Need this, as bulletin service is public
	removeUserInfoFromPosts(c, posts)

	t.RespondOK(c, posts)
}

// CreateDiscussionParams is a struct to hold the parameters used to create a discussion.
type CreateDiscussionParams struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// CreateDiscussion godoc
// @Summary Create discussion
// @Description Creates a discussion
// @ID bulletins-create-discussion
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param createParams body bulletin.CreateDiscussionParams true "Create Discussion Params"
// @Success 201 {object} models.Discussion
// @Failure 1101 {object} lib.APIError "request data validation failed"
// @Failure 1332 {object} lib.APIError "authenticated user not found"
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/discussions [post]
func (t *Controller) CreateDiscussion(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	createData := CreateDiscussionParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	user := new(models.User)
	if err = t.userModel.GetUserByID(userID, user); err != nil {
		// Don't return 404; instead, return authed user not found
		if gorm.IsRecordNotFoundError(err) {
			c.Set(services.UpdateTokenKey, true)
			err = lib.ErrorAuthedUserNotFound
		}

		t.RespondError(c, err)
		return
	}

	// Construct new discussion
	discussion := models.Discussion{
		Title:      createData.Title,
		UserID:     userID,
		ExUserName: user.Name,
		Posts: []*models.Post{
			{
				UserID:     userID,
				ExUserName: user.Name,
				Content:    createData.Content,
			},
		},
	}

	err = t.discussionModel.CreateDiscussion(&discussion)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, discussion)
}

// DeleteDiscussion godoc
// @Summary Delete discussion
// @Description Deletes a discussion (must be admin)
// @ID bulletins-delete-discussion
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param discussionID path uint true "Discussion ID"
// @Success 200 {object} models.Discussion
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/discussion/{discussionID} [delete]
func (t *Controller) DeleteDiscussion(c *gin.Context) {
	// Decode discussionID.
	discussionID, err := services.GetUIntParam(c, "discussionID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query to get discussion
	var discussion models.Discussion
	err = t.discussionModel.GetDiscussionByIDFullPreload(discussionID, &discussion)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Delete discussion
	err = t.discussionModel.DeleteDiscussion(&discussion)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, discussion)
}
