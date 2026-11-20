package user

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// FlagUserPhoto godoc
// @Summary Flag a user's profile photo
// @Description creates a USER_PHOTO_REVIEW moderator task for the user's profile photo
// @ID flagUserPhoto
// @Tags users
// @Accept json
// @Produce json
// @Param userID path uint true "User ID"
// @Success 200
// @Failure 2416 {object} services.BaseErrorResponse "user has no profile photo"
// @Failure 2417 {object} services.BaseErrorResponse "open flag already exists for this photo"
// @Failure 2418 {object} services.BaseErrorResponse "cannot flag your own photo"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /users/{userID}/photo/flag [post]
func (t *Controller) FlagUserPhoto(c *gin.Context) {
	targetUserID, err := services.GetUIntParam(c, "userID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	userID := services.GetUserID(c)
	if targetUserID == userID {
		t.RespondAPIError(c, lib.ErrorUserCannotFlagOwnPhoto)
		return
	}

	var target models.User
	if err := t.userModel.GetUserByID(targetUserID, &target); err != nil {
		t.RespondError(c, err)
		return
	}

	exists, err := t.pictureBackend.DoesUserPhotoExists(target.UnixID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists || (target.PhotoHidden != nil && *target.PhotoHidden) {
		t.RespondAPIError(c, lib.ErrorUserPhotoNotFound)
		return
	}

	task, err := models.NewUserPhotoReviewTask(&userID, targetUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	open, err := t.moderatorTaskModel.HasOpenTaskForRequester(userID, task.TaskReference)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if open {
		t.RespondAPIError(c, lib.ErrorUserPhotoFlagOpen)
		return
	}

	if err := t.moderatorTaskModel.CreateTask(task); err != nil {
		t.RespondError(c, err)
		return
	}

	t.Log.Infow("user photo flagged as moderator task",
		"targetUserID", targetUserID,
		"taskID", task.ID,
		"taskReference", task.TaskReference,
		"requestingUserID", userID,
	)

	t.RespondOK(c, nil)
}
