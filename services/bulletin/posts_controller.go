package bulletin

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GetPost godoc
// @Summary Get post
// @Description Get post by ID with user, posts, and posts.user preloaded
// @ID bulletins-get-post
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param postID path uint true "Post ID"
// @Success 200 {object} models.Post
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/posts/{postID} [get]
func (t *Controller) GetPost(c *gin.Context) {
	// Decode postID.
	postID, err := services.GetUIntParam(c, "postID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var post models.Post
	err = t.postModel.GetPostByID(postID, &post)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Remove user info if not a user. Need this, as bulletin service is public
	if hasUserAuth(c) {
		sanitize.User(post.User, c)
		if post.Discussion != nil {
			sanitize.User(post.Discussion.User, c)
		}
	} else {
		post.User = nil
		post.ExUserName = ""
		if post.Discussion != nil {
			post.Discussion.User = nil
			post.Discussion.ExUserName = ""
		}
	}

	t.RespondOK(c, post)
}

// CreatePostParams is a struct to hold the parameters used to create a post.
type CreatePostParams struct {
	DiscussionID uint   `json:"discussionID" binding:"required"`
	Content      string `json:"content" binding:"required"`
}

// CreatePost godoc
// @Summary Create post
// @Description Creates a post
// @ID bulletins-create-post
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param createParams body bulletin.CreatePostParams true "Create Post Params"
// @Success 201 {object} models.Post
// @Failure 1101 {object} lib.APIError "request data validation failed"
// @Failure 1332 {object} lib.APIError "authenticated user not found"
// @Failure 1850 {object} lib.APIError "discussion cannot be found"
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/posts [post]
func (t *Controller) CreatePost(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	createData := CreatePostParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Discussion must exist
	exists, err := t.discussionModel.DoesDiscussionExist(createData.DiscussionID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorDiscussionNotFound)
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

	// Construct new post
	post := models.Post{
		UserID:       userID,
		DiscussionID: createData.DiscussionID,
		Content:      createData.Content,
		ExUserName:   user.Name,
	}

	err = t.postModel.CreatePostByDiscussion(createData.DiscussionID, &post)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, post)
}

// UpdatePostParams is a struct to hold the parameters used to update a post.
type UpdatePostParams struct {
	Content string `json:"content"`
}

// UpdatePost godoc
// @Summary Update post
// @Description Updates a post
// @ID bulletins-update-post
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param updateParams body bulletin.UpdatePostParams true "Update Post Params"
// @Param postID path uint true "Post ID"
// @Success 200 {object} models.Post
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/posts/{postID} [patch]
func (t *Controller) UpdatePost(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode postID.
	postID, err := services.GetUIntParam(c, "postID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Bind update params
	updateData := UpdatePostParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	var post models.Post
	err = t.postModel.GetPostByID(postID, &post)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Must own post
	if post.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	if updateData.Content != "" {
		post.Content = updateData.Content
	}

	err = t.postModel.UpdatePost(&post)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, post)
}

// DeletePost godoc
// @Summary Delete post
// @Description Deletes a post
// @ID bulletins-delete-post
// @Tags bulletins
// @Accept  json
// @Produce  json
// @Param postID path uint true "Post ID"
// @Success 200 {object} models.Post
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /bulletin/posts/{postID} [delete]
func (t *Controller) DeletePost(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode postID.
	postID, err := services.GetUIntParam(c, "postID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var post models.Post
	err = t.postModel.GetPostByID(postID, &post)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Must own post or be admin
	if post.UserID != userID && !auth.HasScope(c, auth.ScopeAdminAll) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	err = t.postModel.DeletePost(&post)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, post)
}
