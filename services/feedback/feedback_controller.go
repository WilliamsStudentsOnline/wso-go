package feedback

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// CreateFeedback godoc
// @Summary Create feedback
// @Description submit site feedback
// @ID createFeedback
// @Tags feedback
// @Accept json
// @Produce json
// @Param body body feedback.CreateFeedbackParams true "Feedback"
// @Success 201 {object} models.Feedback
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /feedback [post]
func (t *Controller) CreateFeedback(c *gin.Context) {
	var params CreateFeedbackParams
	if err := c.ShouldBindJSON(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	item := models.Feedback{
		PreferredName: params.PreferredName,
		URL:           params.URL,
		Notes:         params.Notes,
		Priority:      models.FeedbackPriorityNone,
	}
	if params.UserID != nil {
		if !auth.CheckIDIsSelf(c, *params.UserID) {
			t.RespondAPIError(c, lib.ErrorMustBeSelf)
			return
		}
		item.UserID = params.UserID
	}

	if err := t.feedbackModel.CreateFeedback(&item); err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondCreated(c, item)
}

// ListFeedback godoc
// @Summary List feedback
// @Description list feedback submissions for admin review
// @ID listFeedback
// @Tags admin,feedback
// @Accept json
// @Produce json
// @Param priority query string false "Filter by priority" Enums(none,priority_low,priority_med,priority_high,resolved,dismissed)
// @Param offset query int false "Offset pagination"
// @Param limit query int false "Limit pagination"
// @Param preload query []string false "Preload" Enums(user)
// @Success 200 {array} models.Feedback
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/feedback [get]
func (t *Controller) ListFeedback(c *gin.Context) {
	params := models.GetAllFeedbackOptions{}
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var items []*models.Feedback
	if err := t.feedbackModel.GetAllFeedback(&items, &params); err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.feedbackModel.CountAllFeedback(&params)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)
	t.RespondOK(c, items)
}

// GetFeedback godoc
// @Summary Get feedback
// @Description get a single feedback submission by id
// @ID getFeedback
// @Tags admin,feedback
// @Accept json
// @Produce json
// @Param feedbackID path uint true "Feedback ID"
// @Param preload query []string false "Preload" Enums(user)
// @Success 200 {object} models.Feedback
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/feedback/{feedbackID} [get]
func (t *Controller) GetFeedback(c *gin.Context) {
	feedbackID, err := services.GetUIntParam(c, "feedbackID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	params := models.GetAllFeedbackOptions{}
	_ = c.ShouldBindQuery(&params)

	var item models.Feedback
	if err := t.feedbackModel.GetFeedbackByID(feedbackID, &item, params.Preload...); err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, item)
}

// UpdateFeedback godoc
// @Summary Update feedback
// @Description update admin notes, priority, or linked issue
// @ID updateFeedback
// @Tags admin,feedback
// @Accept json
// @Produce json
// @Param feedbackID path uint true "Feedback ID"
// @Param body body feedback.UpdateFeedbackParams true "Update"
// @Success 200 {object} models.Feedback
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/feedback/{feedbackID} [patch]
func (t *Controller) UpdateFeedback(c *gin.Context) {
	feedbackID, err := services.GetUIntParam(c, "feedbackID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	var params UpdateFeedbackParams
	if err := c.ShouldBindJSON(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	updates := map[string]interface{}{}
	if params.AdminNotes != nil {
		updates["admin_notes"] = *params.AdminNotes
	}
	if params.Issue != nil {
		updates["issue"] = *params.Issue
	}
	if params.Priority != nil {
		if err := params.Priority.Validate(); err != nil {
			t.RespondBadBind(c, err)
			return
		}
		updates["priority"] = *params.Priority
	}

	if err := t.feedbackModel.UpdateFeedback(feedbackID, updates); err != nil {
		t.RespondError(c, err)
		return
	}

	var item models.Feedback
	if err := t.feedbackModel.GetFeedbackByID(feedbackID, &item); err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, item)
}

// DeleteFeedback godoc
// @Summary Delete feedback
// @Description soft-delete a feedback submission
// @ID deleteFeedback
// @Tags admin,feedback
// @Accept json
// @Produce json
// @Param feedbackID path uint true "Feedback ID"
// @Success 200 {object} map[string]uint
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/feedback/{feedbackID} [delete]
func (t *Controller) DeleteFeedback(c *gin.Context) {
	feedbackID, err := services.GetUIntParam(c, "feedbackID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if err := t.feedbackModel.SoftDeleteFeedbackByID(feedbackID); err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, gin.H{"deleted": feedbackID})
}
