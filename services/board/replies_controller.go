package board

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type CreateReplyParams struct {
	Content string `json:"content" binding:"required"`
}

// CreateReply godoc
// @Summary Create a reply on a thread
// @Tags board
// @Accept json
// @Produce json
// @Param threadID path uint true "Thread ID"
// @Param body body CreateReplyParams true "Reply content"
// @Success 201 {object} models.BoardPost
// @Router /board/threads/{threadID}/replies [post]
func (t *Controller) CreateReply(c *gin.Context) {
	userID := services.GetUserID(c)

	threadID, err := services.GetUIntParam(c, "threadID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var params CreateReplyParams
	if err := c.ShouldBind(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var thread models.BoardThread
	if err := t.threadModel.GetThreadByID(threadID, &thread, &models.GetBoardThreadByIDOptions{
		Preload: []string{},
	}); err != nil {
		t.RespondError(c, err)
		return
	}

	if !thread.RepliesEnabled {
		t.RespondAPIError(c, lib.ErrorBoardRepliesDisabled)
		return
	}

	user := new(models.User)
	if err := t.userModel.GetUserByID(userID, user); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Set(services.UpdateTokenKey, true)
			err = lib.ErrorAuthedUserNotFound
		}
		t.RespondError(c, err)
		return
	}

	post := &models.BoardPost{
		UserID:     userID,
		ExUserName: user.Name,
		Content:    params.Content,
	}

	if err := t.postModel.CreateReply(threadID, post); err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, post)
}

type UpdateReplyParams struct {
	Content *string `json:"content"`
}

// UpdateReply godoc
// @Summary Update a reply
// @Tags board
// @Accept json
// @Produce json
// @Param replyID path uint true "Reply ID"
// @Param body body UpdateReplyParams true "Update params"
// @Success 200 {object} models.BoardPost
// @Router /board/replies/{replyID} [patch]
func (t *Controller) UpdateReply(c *gin.Context) {
	replyID, err := services.GetUIntParam(c, "replyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var params UpdateReplyParams
	if err := c.ShouldBind(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var post models.BoardPost
	if err := t.postModel.GetPostByID(replyID, &post); err != nil {
		t.RespondError(c, err)
		return
	}

	if !canMutate(c, post.UserID) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	if params.Content != nil {
		post.Content = *params.Content
	}

	if err := t.postModel.UpdatePost(&post); err != nil {
		t.RespondError(c, err)
		return
	}

	sanitizePosts(c, []*models.BoardPost{&post})
	t.RespondOK(c, post)
}

// DeleteReply godoc
// @Summary Delete a reply (OP delete cascades to thread)
// @Tags board
// @Param replyID path uint true "Reply ID"
// @Success 200 {object} models.BoardPost
// @Router /board/replies/{replyID} [delete]
func (t *Controller) DeleteReply(c *gin.Context) {
	replyID, err := services.GetUIntParam(c, "replyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var post models.BoardPost
	if err := t.postModel.GetPostByID(replyID, &post); err != nil {
		t.RespondError(c, err)
		return
	}

	if !canMutate(c, post.UserID) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	if err := t.postModel.DeletePost(&post); err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, post)
}
