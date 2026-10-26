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
// @Description lists moderator tasks grouped by subjectKey (default: open only)
// @ID listModeratorTasks
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskType query string false "Filter by task type" Enums(FACTRAK_REVIEW)
// @Param resolved query bool false "If true, closed tasks; if false/omitted, open tasks"
// @Param subjectKey query string false "Filter by subject key"
// @Param offset query int false "Offset pagination"
// @Param limit query int false "Limit pagination"
// @Param preload query []string false "Preload" Enums(requestingUser,resolvedByUser)
// @Success 200 {array} admin.ModeratorTaskGroupDTO
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
			t.Log.Warnw("moderator task context enrich failed", "subjectKey", g.SubjectKey, "err", err)
		}
	}

	t.RespondOK(c, groups)
}

// GetTask godoc
// @Summary Get moderator task
// @Description get a single moderator task by id with enrich-on-read context
// @ID getModeratorTask
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskID path uint true "Task ID"
// @Param preload query []string false "Preload" Enums(requestingUser,resolvedByUser)
// @Success 200 {object} admin.ModeratorTaskGroupDTO
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

	group := &ModeratorTaskGroupDTO{
		SubjectKey: task.SubjectKey,
		TaskType:   task.TaskType,
		Tasks:      []*ModeratorTaskDTO{taskToDTO(&task)},
	}
	if err := t.attachContext(group); err != nil {
		t.Log.Warnw("moderator task context enrich failed", "subjectKey", group.SubjectKey, "err", err)
	}

	t.RespondOK(c, group)
}

// GetSubjectKey godoc
// @Summary Compute moderator task subject key
// @Description returns SubjectKey(taskType, id...) — pass identifying ids as repeated id query params
// @ID getModeratorTaskSubjectKey
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param taskType query string true "Task type" Enums(FACTRAK_REVIEW)
// @Param id query []int true "Identifying id(s) for the subject"
// @Success 200 {object} admin.SubjectKeyResponse
// @Failure 400 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/subject-key [get]
func (t *Controller) GetSubjectKey(c *gin.Context) {
	taskType := models.ModeratorTaskType(c.Query("taskType"))
	if err := taskType.Validate(); err != nil {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidType)
		return
	}

	rawIDs := c.QueryArray("id")
	if len(rawIDs) == 0 {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidSubjectKey)
		return
	}
	ids := make([]uint, len(rawIDs))
	for i, raw := range rawIDs {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidSubjectKey)
			return
		}
		ids[i] = uint(id)
	}

	t.RespondOK(c, SubjectKeyResponse{SubjectKey: models.SubjectKey(taskType, ids...)})
}

// ResolveTasksBySubject godoc
// @Summary Resolve moderator tasks by subject
// @Description approve or reject all open tasks sharing a subjectKey; approve runs the type-specific side effect once
// @ID resolveModeratorTasksBySubject
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param subjectKey path string true "Subject key (e.g. FACTRAK_REVIEW:123)"
// @Param body body admin.ResolveTasksBody true "Resolution"
// @Success 200 {object} admin.ModeratorTaskGroupDTO
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 409 {object} services.BaseErrorResponse "subject already gone — discard open tasks"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/by-subject/{subjectKey}/resolve [post]
func (t *Controller) ResolveTasksBySubject(c *gin.Context) {
	subjectKey := c.Param("subjectKey")
	if subjectKey == "" {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidSubjectKey)
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
	if err := t.moderatorTaskModel.GetOpenTasksBySubjectKey(subjectKey, &open); err != nil {
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
		"subjectKey", subjectKey,
		"taskType", taskType,
		"resolution", body.Resolution,
		"actorUserID", actorID,
		"openTaskCount", len(open),
	)

	if body.Resolution == models.ModeratorTaskResolutionApproved {
		if err := t.applyApproveSideEffect(c, taskType, open[0]); err != nil {
			t.Log.Warnw("moderator approve side effect failed",
				"subjectKey", subjectKey,
				"taskType", taskType,
				"actorUserID", actorID,
				"err", err,
			)
			t.RespondError(c, err)
			return
		}
	}

	now := time.Now().UTC()
	n, err := t.moderatorTaskModel.ResolveOpenTasksBySubjectKey(subjectKey, body.Resolution, actorID, now)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.Log.Infow("moderator resolve completed",
		"subjectKey", subjectKey,
		"taskType", taskType,
		"resolution", body.Resolution,
		"actorUserID", actorID,
		"rowsUpdated", n,
	)

	var updated []*models.ModeratorTask
	falseVal := false
	opts := &models.GetAllModeratorTasksOptions{
		SubjectKey: &subjectKey,
		Resolved:   &falseVal,
	}
	// Re-fetch including just-resolved: query with resolved=true for this subject
	trueVal := true
	opts.Resolved = &trueVal
	if err := t.moderatorTaskModel.GetAllTasks(&updated, opts); err != nil {
		t.RespondError(c, err)
		return
	}

	groups := groupTasks(updated)
	var group *ModeratorTaskGroupDTO
	if len(groups) > 0 {
		group = groups[0]
	} else {
		group = &ModeratorTaskGroupDTO{SubjectKey: subjectKey, TaskType: taskType, Tasks: []*ModeratorTaskDTO{}}
	}
	_ = t.attachContext(group)

	t.RespondOK(c, group)
}

// DiscardTasksBySubject godoc
// @Summary Discard open moderator tasks by subject
// @Description soft-deletes all open tasks for a subjectKey (cleanup when subject is already gone)
// @ID discardModeratorTasksBySubject
// @Tags admin,moderator
// @Accept json
// @Produce json
// @Param subjectKey path string true "Subject key"
// @Success 200 {object} map[string]int64
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/tasks/by-subject/{subjectKey} [delete]
func (t *Controller) DiscardTasksBySubject(c *gin.Context) {
	subjectKey := c.Param("subjectKey")
	if subjectKey == "" {
		t.RespondAPIError(c, lib.ErrorModeratorTaskInvalidSubjectKey)
		return
	}

	actorID := services.GetUserID(c)
	n, err := t.moderatorTaskModel.SoftDeleteOpenTasksBySubjectKey(subjectKey)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if n == 0 {
		t.RespondAPIError(c, lib.ErrorModeratorTaskNoOpenTasks)
		return
	}

	t.Log.Infow("moderator tasks discarded",
		"subjectKey", subjectKey,
		"actorUserID", actorID,
		"rowsDeleted", n,
	)

	t.RespondOK(c, gin.H{"discarded": n})
}

func (t *Controller) attachContext(group *ModeratorTaskGroupDTO) error {
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
		// Survey is already deleted; do not fail the resolve on deficit bookkeeping.
	}

	c.Set(services.UpdateTokenKey, true)
	t.Log.Infow("factrak survey deleted via moderator approve",
		"surveyID", fields.FactrakSurveyID,
		"subjectKey", sample.SubjectKey,
		"actorUserID", services.GetUserID(c),
	)
	return nil
}
