package board

import (
	"net/http"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type listThreadsResponse struct {
	Threads    []*models.BoardThread `json:"threads"`
	NextCursor string                `json:"nextCursor,omitempty"`
}

// ListThreads godoc
// @Summary List board threads
// @Tags board
// @Produce json
// @Param cursor query string false "Cursor pagination"
// @Param limit query int false "Limit"
// @Param sort query string false "activity|created|startsAt"
// @Param includeNull query bool false "Include null startsAt when sorting by startsAt"
// @Param type[] query []string false "Thread types"
// @Param resolved query bool false "Filter by resolved"
// @Param offeringRide query bool false "Filter rides by offeringRide"
// @Param userID query int false "Filter by author"
// @Success 200 {object} listThreadsResponse
// @Router /board/threads [get]
func (t *Controller) ListThreads(c *gin.Context) {
	params := models.GetBoardThreadsOptions{}
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var threads []*models.BoardThread
	if err := t.threadModel.GetThreads(&threads, &params); err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.threadModel.CountThreads(&params)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	sanitizeThreads(c, threads)

	resp := listThreadsResponse{Threads: threads}
	if len(threads) > 0 {
		sort := params.Sort
		if sort == "" {
			sort = "activity"
		}
		resp.NextCursor = models.BoardCursorForThread(threads[len(threads)-1], sort)
	}

	t.RespondOK(c, resp)
}

// GetThread godoc
// @Summary Get board thread
// @Tags board
// @Produce json
// @Param threadID path uint true "Thread ID"
// @Success 200 {object} models.BoardThread
// @Router /board/threads/{threadID} [get]
func (t *Controller) GetThread(c *gin.Context) {
	threadID, err := services.GetUIntParam(c, "threadID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	params := models.GetBoardThreadByIDOptions{}
	_ = c.ShouldBindQuery(&params)

	var thread models.BoardThread
	if err := t.threadModel.GetThreadByID(threadID, &thread, &params); err != nil {
		t.RespondError(c, err)
		return
	}

	sanitizeThread(c, &thread)
	t.RespondOK(c, thread)
}

type createThreadRideParams struct {
	OfferingRide *bool  `json:"offeringRide"`
	Source       string `json:"source"`
	Destination  string `json:"destination"`
}

type CreateThreadParams struct {
	Type           string                  `json:"type" binding:"required"`
	Title          string                  `json:"title" binding:"required"`
	Body           string                  `json:"body" binding:"required"`
	RepliesEnabled *bool                   `json:"repliesEnabled"`
	Resolved       *bool                   `json:"resolved"`
	StartsAt       *time.Time              `json:"startsAt"`
	EndsAt         *time.Time              `json:"endsAt"`
	Ride           *createThreadRideParams `json:"ride"`
}

// CreateThread godoc
// @Summary Create board thread
// @Tags board
// @Accept json
// @Produce json
// @Param body body CreateThreadParams true "Create params"
// @Success 201 {object} models.BoardThread
// @Router /board/threads [post]
func (t *Controller) CreateThread(c *gin.Context) {
	userID := services.GetUserID(c)

	var params CreateThreadParams
	if err := c.ShouldBind(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if !models.IsValidBoardThreadType(params.Type) {
		t.RespondAPIError(c, lib.ErrorBoardInvalidType)
		return
	}

	if params.StartsAt != nil && params.EndsAt != nil && params.StartsAt.After(*params.EndsAt) {
		t.RespondAPIError(c, lib.ErrorBoardInvalidDates)
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

	repliesEnabled := true
	if params.RepliesEnabled != nil {
		repliesEnabled = *params.RepliesEnabled
	}

	thread := &models.BoardThread{
		Type:           params.Type,
		Title:          params.Title,
		UserID:         userID,
		ExUserName:     user.Name,
		RepliesEnabled: repliesEnabled,
		Resolved:       params.Resolved,
		StartsAt:       params.StartsAt,
		EndsAt:         params.EndsAt,
		LastActive:     time.Now(),
	}

	var ride *models.BoardRideMeta
	if params.Type == models.BoardThreadTypeRide {
		if params.Ride == nil || params.Ride.OfferingRide == nil ||
			params.Ride.Source == "" || params.Ride.Destination == "" {
			t.RespondAPIError(c, lib.ErrorBoardRideMetaRequired)
			return
		}
		ride = &models.BoardRideMeta{
			OfferingRide: *params.Ride.OfferingRide,
			Source:       params.Ride.Source,
			Destination:  params.Ride.Destination,
		}
	}

	if err := t.threadModel.CreateThread(thread, params.Body, ride); err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, thread)
}

type UpdateThreadParams struct {
	Title          *string                 `json:"title"`
	Body           *string                 `json:"body"`
	RepliesEnabled *bool                   `json:"repliesEnabled"`
	Resolved       *bool                   `json:"resolved"`
	StartsAt       *time.Time              `json:"startsAt"`
	EndsAt         *time.Time              `json:"endsAt"`
	Type           *string                 `json:"type"`
	Ride           *createThreadRideParams `json:"ride"`
}

// UpdateThread godoc
// @Summary Update board thread
// @Tags board
// @Accept json
// @Produce json
// @Param threadID path uint true "Thread ID"
// @Param body body UpdateThreadParams true "Update params"
// @Success 200 {object} models.BoardThread
// @Router /board/threads/{threadID} [patch]
func (t *Controller) UpdateThread(c *gin.Context) {
	userID := services.GetUserID(c)

	threadID, err := services.GetUIntParam(c, "threadID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var params UpdateThreadParams
	if err := c.ShouldBind(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if params.Type != nil {
		t.RespondAPIError(c, lib.ErrorBoardTypeImmutable)
		return
	}

	var thread models.BoardThread
	if err := t.threadModel.GetThreadByID(threadID, &thread, &models.GetBoardThreadByIDOptions{
		Preload: []string{"user", "ride", "posts"},
	}); err != nil {
		t.RespondError(c, err)
		return
	}

	if !canMutate(c, thread.UserID) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	if params.Title != nil {
		thread.Title = *params.Title
	}
	if params.RepliesEnabled != nil {
		thread.RepliesEnabled = *params.RepliesEnabled
	}
	if params.Resolved != nil {
		thread.Resolved = params.Resolved
	}
	if params.StartsAt != nil {
		thread.StartsAt = params.StartsAt
	}
	if params.EndsAt != nil {
		thread.EndsAt = params.EndsAt
	}
	if thread.StartsAt != nil && thread.EndsAt != nil && thread.StartsAt.After(*thread.EndsAt) {
		t.RespondAPIError(c, lib.ErrorBoardInvalidDates)
		return
	}

	if err := t.threadModel.UpdateThread(&thread); err != nil {
		t.RespondError(c, err)
		return
	}

	if params.Body != nil {
		if err := t.threadModel.UpdateOPContent(thread.ID, *params.Body); err != nil {
			t.RespondError(c, err)
			return
		}
		thread.Body = *params.Body
	}

	if params.Ride != nil && thread.Type == models.BoardThreadTypeRide {
		if thread.RideMeta == nil {
			thread.RideMeta = &models.BoardRideMeta{ThreadID: thread.ID}
		}
		if params.Ride.OfferingRide != nil {
			thread.RideMeta.OfferingRide = *params.Ride.OfferingRide
		}
		if params.Ride.Source != "" {
			thread.RideMeta.Source = params.Ride.Source
		}
		if params.Ride.Destination != "" {
			thread.RideMeta.Destination = params.Ride.Destination
		}
		if err := t.threadModel.DB.Save(thread.RideMeta).Error; err != nil {
			t.RespondError(c, err)
			return
		}
	}

	_ = userID // used via canMutate
	sanitizeThread(c, &thread)
	t.RespondOK(c, thread)
}

// DeleteThread godoc
// @Summary Delete board thread
// @Tags board
// @Param threadID path uint true "Thread ID"
// @Success 200 {object} models.BoardThread
// @Router /board/threads/{threadID} [delete]
func (t *Controller) DeleteThread(c *gin.Context) {
	threadID, err := services.GetUIntParam(c, "threadID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var thread models.BoardThread
	if err := t.threadModel.GetThreadByID(threadID, &thread, &models.GetBoardThreadByIDOptions{
		Preload: []string{"user", "ride", "posts"},
	}); err != nil {
		t.RespondError(c, err)
		return
	}

	if !canMutate(c, thread.UserID) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	if err := t.threadModel.DeleteThread(&thread); err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, thread)
}
