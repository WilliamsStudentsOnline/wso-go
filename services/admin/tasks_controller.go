package admin

import (
	"encoding/json"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// ListTasks godoc
// @Summary List moderator tasks
// @Description lists moderator tasks grouped by taskReference
// @ID listModeratorTasks
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskType query string false "Filter by task type" Enums(FACTRAK_REVIEW,NEW_USER_TAG,DELETE_USER_TAG)
// @Param resolvedAt query string false "Omit for open tasks only; set any time to list resolved tasks"
// @Param taskReference query string false "Filter by task reference"
// @Param offset query int false "Offset pagination"
// @Param limit query int false "Limit pagination"
// @Param preload query []string false "Preload" Enums(requestingUser,resolvedByUser)
// @Success 200 {array} admin.ModeratorTaskGroup
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks [get]
func (t *Controller) ListTasks(c *gin.Context) {
	params := models.GetAllModeratorTasksOptions{}
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var tasks []*models.ModeratorTask
	if err := t.moderatorTaskModel.GetAllTasks(&tasks, &params); err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.moderatorTaskModel.CountAllTasks(&params)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	groups := groupTasks(tasks)
	for _, g := range groups {
		if err := t.attachGroupContext(g); err != nil {
			t.Log.Warnw("task context enrich failed", "taskReference", g.TaskReference, "err", err)
		}
	}

	t.RespondOK(c, groups)
}

// GetTask godoc
// @Summary Get moderator task
// @Description get a single moderator task by id
// @ID getModeratorTask
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskID path uint true "Task ID"
// @Param preload query []string false "Preload" Enums(requestingUser,resolvedByUser)
// @Success 200 {object} admin.ModeratorTaskResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/{taskID} [get]
func (t *Controller) GetTask(c *gin.Context) {
	taskID, err := services.GetUIntParam(c, "taskID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	params := models.GetAllModeratorTasksOptions{}
	_ = c.ShouldBindQuery(&params)

	var task models.ModeratorTask
	if err := t.moderatorTaskModel.GetTaskByID(taskID, &task, params.Preload...); err != nil {
		t.RespondError(c, err)
		return
	}

	resp := taskToResponse(&task)
	ctx, err := t.contextForTask(task.TaskType, task.TaskFields)
	if err != nil {
		t.Log.Warnw("task context enrich failed", "taskID", task.ID, "err", err)
	} else {
		resp.Context = ctx
	}

	t.RespondOK(c, resp)
}

// ResolveTasks godoc
// @Summary Resolve moderator tasks
// @Description approve or reject all open tasks that share this task's reference
// @ID resolveModeratorTasks
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskID path uint true "Task ID"
// @Param body body admin.ResolveTasksBody true "Resolution"
// @Success 200 {object} admin.ModeratorTaskGroup
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 409 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/{taskID}/resolve [post]
func (t *Controller) ResolveTasks(c *gin.Context) {
	taskID, err := services.GetUIntParam(c, "taskID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	var body ResolveTasksBody
	if err := c.ShouldBindJSON(&body); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	if body.Resolution != models.ModeratorTaskResolutionApproved &&
		body.Resolution != models.ModeratorTaskResolutionRejected {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidResolution)
		return
	}

	var seed models.ModeratorTask
	if err := t.moderatorTaskModel.GetTaskByID(taskID, &seed); err != nil {
		t.RespondError(c, err)
		return
	}

	taskReference := seed.TaskReference
	var open []*models.ModeratorTask
	if err := t.moderatorTaskModel.GetOpenTasksByTaskReference(taskReference, &open); err != nil {
		t.RespondError(c, err)
		return
	}
	if len(open) == 0 {
		t.RespondAPIError(c, lib.ErrorModeratorTaskNoOpenTasks)
		return
	}

	actorID := services.GetUserID(c)
	taskType := open[0].TaskType

	t.Log.Infow("moderator resolve requested",
		"taskID", taskID,
		"taskReference", taskReference,
		"taskType", taskType,
		"resolution", body.Resolution,
		"actorUserID", actorID,
		"openTaskCount", len(open),
	)

	if body.Resolution == models.ModeratorTaskResolutionApproved {
		if err := t.applyApproveSideEffect(c, taskType, open[0]); err != nil {
			t.Log.Warnw("moderator approve side effect failed",
				"taskID", taskID,
				"taskReference", taskReference,
				"taskType", taskType,
				"actorUserID", actorID,
				"err", err,
			)
			t.RespondError(c, err)
			return
		}
	}

	now := time.Now().UTC()
	n, err := t.moderatorTaskModel.ResolveOpenTasksByTaskReference(taskReference, body.Resolution, actorID, now)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.Log.Infow("moderator resolve completed",
		"taskID", taskID,
		"taskReference", taskReference,
		"taskType", taskType,
		"resolution", body.Resolution,
		"actorUserID", actorID,
		"rowsUpdated", n,
	)

	trueVal := now
	opts := &models.GetAllModeratorTasksOptions{
		TaskReference: &taskReference,
		ResolvedAt:    &trueVal,
	}
	var updated []*models.ModeratorTask
	if err := t.moderatorTaskModel.GetAllTasks(&updated, opts); err != nil {
		t.RespondError(c, err)
		return
	}

	groups := groupTasks(updated)
	var group *ModeratorTaskGroup
	if len(groups) > 0 {
		group = groups[0]
	} else {
		group = &ModeratorTaskGroup{TaskReference: taskReference, TaskType: taskType, Tasks: []*ModeratorTaskResponse{}}
	}
	_ = t.attachGroupContext(group)

	t.RespondOK(c, group)
}

// DeleteTask godoc
// @Summary Delete moderator task
// @Description soft-deletes a single moderator task by id
// @ID deleteModeratorTask
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskID path uint true "Task ID"
// @Success 200 {object} map[string]uint
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/{taskID} [delete]
func (t *Controller) DeleteTask(c *gin.Context) {
	taskID, err := services.GetUIntParam(c, "taskID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if err := t.moderatorTaskModel.SoftDeleteTaskByID(taskID); err != nil {
		t.RespondError(c, err)
		return
	}

	t.Log.Infow("moderator task deleted",
		"taskID", taskID,
		"actorUserID", services.GetUserID(c),
	)

	t.RespondOK(c, gin.H{"deleted": taskID})
}

func (t *Controller) attachGroupContext(group *ModeratorTaskGroup) error {
	if len(group.Tasks) == 0 {
		return nil
	}
	raw, err := json.Marshal(group.Tasks[0].TaskFields)
	if err != nil {
		return err
	}
	ctx, err := t.contextForTask(group.TaskType, string(raw))
	if err != nil {
		return err
	}
	group.Context = ctx
	return nil
}

func (t *Controller) contextForTask(taskType models.ModeratorTaskType, taskFieldsJSON string) (interface{}, error) {
	switch taskType {
	case models.ModeratorTaskTypeFactrakReview:
		fields, err := models.DecodeFactrakReviewFields(taskFieldsJSON)
		if err != nil {
			return nil, err
		}
		var survey models.FactrakSurvey
		err = t.surveyModel.GetSurveyByID(fields.FactrakSurveyID, &survey)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return FactrakReviewTaskContext{Survey: nil}, nil
			}
			return nil, err
		}
		anonymizeSurveyAuthor(&survey)
		return FactrakReviewTaskContext{Survey: &survey}, nil
	default:
		return nil, nil
	}
}

func (t *Controller) applyApproveSideEffect(c *gin.Context, taskType models.ModeratorTaskType, sample *models.ModeratorTask) error {
	switch taskType {
	case models.ModeratorTaskTypeFactrakReview:
		return t.approveFactrakReview(c, sample)
	default:
		return lib.ErrorModeratorTaskInvalidType
	}
}

func (t *Controller) approveFactrakReview(c *gin.Context, sample *models.ModeratorTask) error {
	fields, err := models.DecodeFactrakReviewFields(sample.TaskFields)
	if err != nil {
		return err
	}

	var survey models.FactrakSurvey
	err = t.surveyModel.GetSurveyByID(fields.FactrakSurveyID, &survey)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return lib.ErrorModeratorTaskSubjectGone
		}
		return err
	}

	authorID := survey.UserID
	if err := t.surveyModel.DeleteSurvey(&survey); err != nil {
		return err
	}

	user := &models.User{}
	user.ID = authorID
	if err := t.studentModel.UpdateFactrakSurveyDeficit(user); err != nil {
		t.Log.Warnw("factrak survey deficit update after moderator delete failed",
			"surveyID", fields.FactrakSurveyID,
			"authorUserID", authorID,
			"err", err,
		)
	}

	c.Set(services.UpdateTokenKey, true)
	t.Log.Infow("factrak survey deleted via moderator approve",
		"surveyID", fields.FactrakSurveyID,
		"taskReference", sample.TaskReference,
		"actorUserID", services.GetUserID(c),
	)
	return nil
}
