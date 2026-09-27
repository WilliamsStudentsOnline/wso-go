package board

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// FlagThread godoc
// @Summary Flag a board thread
// @Tags board
// @Param threadID path uint true "Thread ID"
// @Success 200
// @Router /board/threads/{threadID}/flag [post]
func (t *Controller) FlagThread(c *gin.Context) {
	t.flagTarget(c, models.FlagTargetBoardThread, "threadID")
}

// FlagReply godoc
// @Summary Flag a board reply
// @Tags board
// @Param replyID path uint true "Reply ID"
// @Success 200
// @Router /board/replies/{replyID}/flag [post]
func (t *Controller) FlagReply(c *gin.Context) {
	t.flagTarget(c, models.FlagTargetBoardPost, "replyID")
}

func (t *Controller) flagTarget(c *gin.Context, targetType, paramName string) {
	userID := services.GetUserID(c)

	targetID, err := services.GetUIntParam(c, paramName)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var ownerID uint
	switch targetType {
	case models.FlagTargetBoardThread:
		var thread models.BoardThread
		if err := t.threadModel.GetThreadByID(targetID, &thread, &models.GetBoardThreadByIDOptions{
			Preload: []string{},
		}); err != nil {
			t.RespondError(c, err)
			return
		}
		ownerID = thread.UserID
	case models.FlagTargetBoardPost:
		var post models.BoardPost
		if err := t.postModel.GetPostByID(targetID, &post); err != nil {
			t.RespondError(c, err)
			return
		}
		ownerID = post.UserID
	}

	if ownerID == userID {
		t.RespondAPIError(c, lib.ErrorBoardCannotFlagSelf)
		return
	}

	flag := &models.Flag{
		TargetType: targetType,
		TargetID:   targetID,
		UserID:     userID,
	}

	created, err := t.flagModel.CreateFlag(flag)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !created {
		t.RespondAPIError(c, lib.ErrorBoardAlreadyFlagged)
		return
	}

	t.RespondOK(c, nil)
}

// ListFlags godoc
// @Summary List flags (mod)
// @Tags board
// @Param targetType query string false "Filter by target type"
// @Success 200 {array} models.Flag
// @Router /board/mod/flags [get]
func (t *Controller) ListFlags(c *gin.Context) {
	params := models.GetFlagsOptions{}
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var flags []*models.Flag
	if err := t.flagModel.ListFlags(&flags, &params); err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.flagModel.CountFlags(&params)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	t.RespondOK(c, flags)
}

type ClearFlagsParams struct {
	TargetType string `json:"targetType" form:"targetType" binding:"required"`
	TargetID   uint   `json:"targetID" form:"targetID" binding:"required"`
}

// ClearFlags godoc
// @Summary Clear flags for a target (mod)
// @Tags board
// @Param targetType query string true "Target type"
// @Param targetID query int true "Target ID"
// @Success 200
// @Router /board/mod/flags [delete]
func (t *Controller) ClearFlags(c *gin.Context) {
	var params ClearFlagsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		// also allow JSON body
		if err2 := c.ShouldBindJSON(&params); err2 != nil {
			t.RespondBadBind(c, err)
			return
		}
	}

	if err := t.flagModel.ClearFlagsForTarget(params.TargetType, params.TargetID); err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
