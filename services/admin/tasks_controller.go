package admin

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// ListTasks godoc
// @Summary List moderator tasks
// @Description lists moderator tasks
// @ID listModeratorTasks
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskType query string false "Filter by task type" Enums(FACTRAK_REVIEW)
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
		if err := t.attachContext(g); err != nil {
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
// @Success 200 {object} admin.ModeratorTaskGroup
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

	group := &ModeratorTaskGroup{
		TaskReference: task.TaskReference,
		TaskType:      task.TaskType,
		Tasks:         []*ModeratorTaskResponse{taskToResponse(&task)},
	}
	if err := t.attachContext(group); err != nil {
		t.Log.Warnw("task context enrich failed", "taskReference", group.TaskReference, "err", err)
	}

	t.RespondOK(c, group)
}

// GetTaskReference godoc
// @Summary Compute moderator task reference
// @Description returns task-specific identifier for which object the task is referring to so we can group requests from multiple users
// @ID getModeratorTaskReference
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskType query string true "Task type" Enums(FACTRAK_REVIEW)
// @Param id query []int true "Identifying id(s) for the reference"
// @Success 200 {object} admin.TaskReferenceResponse
// @Failure 400 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/task-reference [get]
func (t *Controller) GetTaskReference(c *gin.Context) {
	taskType := models.ModeratorTaskType(c.Query("taskType"))
	if err := taskType.Validate(); err != nil {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidType)
		return
	}

	rawIDs := c.QueryArray("id")
	if len(rawIDs) == 0 {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidReference)
		return
	}
	ids := make([]uint, len(rawIDs))
	for i, raw := range rawIDs {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidReference)
			return
		}
		ids[i] = uint(id)
	}

	t.RespondOK(c, TaskReferenceResponse{TaskReference: models.TaskReference(taskType, ids...)})
}

// ResolveTasksByReference godoc
// @Summary Resolve moderator tasks by reference
// @Description approve or reject all open tasks sharing a taskReference with an optional side effect
// @ID resolveModeratorTasksByReference
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskReference path string true "Task reference"
// @Param body body admin.ResolveTasksBody true "Resolution"
// @Success 200 {object} admin.ModeratorTaskGroup
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 409 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/by-reference/{taskReference}/resolve [post]
func (t *Controller) ResolveTasksByReference(c *gin.Context) {
	taskReference := c.Param("taskReference")
	if taskReference == "" {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidReference)
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
		"taskReference", taskReference,
		"taskType", taskType,
		"resolution", body.Resolution,
		"actorUserID", actorID,
		"openTaskCount", len(open),
	)

	if body.Resolution == models.ModeratorTaskResolutionApproved {
		if err := t.applyApproveSideEffect(c, taskType, open[0]); err != nil {
			t.Log.Warnw("moderator approve side effect failed",
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
	_ = t.attachContext(group)

	t.RespondOK(c, group)
}

// DiscardTasksByReference godoc
// @Summary Discard open moderator tasks by reference
// @Description soft-deletes all open tasks for a taskReference
// @ID discardModeratorTasksByReference
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskReference path string true "Task reference"
// @Success 200 {object} map[string]int64
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/by-reference/{taskReference} [delete]
func (t *Controller) DiscardTasksByReference(c *gin.Context) {
	taskReference := c.Param("taskReference")
	if taskReference == "" {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidReference)
		return
	}

	actorID := services.GetUserID(c)
	n, err := t.moderatorTaskModel.SoftDeleteOpenTasksByTaskReference(taskReference)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if n == 0 {
		t.RespondAPIError(c, lib.ErrorModeratorTaskNoOpenTasks)
		return
	}

	t.Log.Infow("moderator tasks discarded",
		"taskReference", taskReference,
		"actorUserID", actorID,
		"rowsDeleted", n,
	)

	t.RespondOK(c, gin.H{"discarded": n})
}

func (t *Controller) attachContext(group *ModeratorTaskGroup) error {
	switch group.TaskType {
	case models.ModeratorTaskTypeFactrakReview:
		if len(group.Tasks) == 0 {
			return nil
		}
		raw, err := json.Marshal(group.Tasks[0].TaskFields)
		if err != nil {
			return err
		}
		fields, err := models.DecodeFactrakReviewFields(string(raw))
		if err != nil {
			return err
		}
		var survey models.FactrakSurvey
		err = t.surveyModel.GetSurveyByID(fields.FactrakSurveyID, &survey)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				group.Context = FactrakReviewTaskContext{Survey: nil}
				return nil
			}
			return err
		}
		anonymizeSurveyAuthor(&survey)
		group.Context = FactrakReviewTaskContext{Survey: &survey}
		return nil
	default:
		return nil
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
