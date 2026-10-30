package board

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// FlagThread godoc
// @Summary Flag a board thread
// @Description creates a BOARD_THREAD_REVIEW moderator task for the thread
// @ID flagBoardThread
// @Tags board
// @Accept json
// @Produce json
// @Param threadID path uint true "Thread ID"
// @Success 200
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 403 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /board/threads/{threadID}/flag [post]
func (t *Controller) FlagThread(c *gin.Context) {
	threadID, err := services.GetUIntParam(c, "threadID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	var thread models.BoardThread
	if err := t.threadModel.GetThreadByID(threadID, &thread, nil); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			t.RespondAPIError(c, lib.ErrorBoardThreadNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	userID := services.GetUserID(c)
	if thread.UserID == userID {
		t.RespondAPIError(c, lib.ErrorBoardCannotFlagSelf)
		return
	}

	task, err := models.NewBoardThreadReviewTask(&userID, threadID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if err := t.moderatorTaskModel.CreateTask(task); err != nil {
		t.RespondError(c, err)
		return
	}

	t.Log.Infow("board thread flagged as moderator task",
		"threadID", threadID,
		"taskID", task.ID,
		"taskReference", task.TaskReference,
		"requestingUserID", userID,
	)

	t.RespondOK(c, nil)
}

// FlagReply godoc
// @Summary Flag a board reply
// @Description creates a BOARD_POST_REVIEW moderator task for the post
// @ID flagBoardReply
// @Tags board
// @Accept json
// @Produce json
// @Param replyID path uint true "Reply ID"
// @Success 200
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 403 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /board/replies/{replyID}/flag [post]
func (t *Controller) FlagReply(c *gin.Context) {
	replyID, err := services.GetUIntParam(c, "replyID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	var post models.BoardPost
	if err := t.postModel.GetPostByID(replyID, &post); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			t.RespondAPIError(c, lib.ErrorBoardPostNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	userID := services.GetUserID(c)
	if post.UserID == userID {
		t.RespondAPIError(c, lib.ErrorBoardCannotFlagSelf)
		return
	}

	task, err := models.NewBoardPostReviewTask(&userID, replyID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if err := t.moderatorTaskModel.CreateTask(task); err != nil {
		t.RespondError(c, err)
		return
	}

	t.Log.Infow("board post flagged as moderator task",
		"postID", replyID,
		"taskID", task.ID,
		"taskReference", task.TaskReference,
		"requestingUserID", userID,
	)

	t.RespondOK(c, nil)
}
