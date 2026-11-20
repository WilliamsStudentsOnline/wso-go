package board

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

const (
	minPollOptions      = 2
	maxPollOptions      = 10
	maxPollOptionLength = 255
)

type createThreadPollParams struct {
	Options  []string   `json:"options"`
	ClosesAt *time.Time `json:"closesAt"`
}

func newPoll(params *createThreadPollParams) (*models.BoardPoll, bool) {
	if len(params.Options) < minPollOptions || len(params.Options) > maxPollOptions {
		return nil, false
	}
	if params.ClosesAt != nil && !params.ClosesAt.After(time.Now()) {
		return nil, false
	}

	poll := &models.BoardPoll{ClosesAt: params.ClosesAt}
	for i, label := range params.Options {
		label = strings.TrimSpace(label)
		if label == "" || utf8.RuneCountInString(label) > maxPollOptionLength {
			return nil, false
		}
		poll.Options = append(poll.Options, &models.BoardPollOption{Position: i, Label: label})
	}
	return poll, true
}

func (t *Controller) fillPollResults(c *gin.Context, poll *models.BoardPoll) error {
	if poll == nil {
		return nil
	}

	if userID := services.GetUserID(c); userID != 0 {
		optionID, err := t.pollModel.GetVoteOptionID(poll.ID, userID)
		if err != nil {
			return err
		}
		poll.MyVote = optionID
	}
	if poll.MyVote == nil && !poll.Closed {
		return nil
	}

	counts, err := t.pollModel.CountVotes(poll.ID)
	if err != nil {
		return err
	}
	total := 0
	for _, o := range poll.Options {
		votes := counts[o.ID]
		o.Votes = &votes
		total += votes
	}
	poll.TotalVotes = &total
	return nil
}

type VoteParams struct {
	OptionID uint `json:"optionID" binding:"required"`
}

// Vote godoc
// @Summary Vote in a thread's poll, replacing any earlier vote
// @ID voteBoardPoll
// @Tags board
// @Accept json
// @Produce json
// @Param threadID path uint true "Thread ID"
// @Param body body VoteParams true "Vote params"
// @Success 200 {object} models.BoardPoll
// @Security Bearer
// @Router /board/threads/{threadID}/vote [put]
func (t *Controller) Vote(c *gin.Context) {
	threadID, err := services.GetUIntParam(c, "threadID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var params VoteParams
	if err := c.ShouldBind(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var poll models.BoardPoll
	if err := t.pollModel.GetPollByThreadID(threadID, &poll); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			t.RespondAPIError(c, lib.ErrorBoardPollNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	if poll.Closed {
		t.RespondAPIError(c, lib.ErrorBoardPollClosed)
		return
	}

	validOption := false
	for _, o := range poll.Options {
		if o.ID == params.OptionID {
			validOption = true
			break
		}
	}
	if !validOption {
		t.RespondAPIError(c, lib.ErrorBoardInvalidPollOption)
		return
	}

	if err := t.pollModel.SetVote(poll.ID, services.GetUserID(c), params.OptionID); err != nil {
		t.RespondError(c, err)
		return
	}

	if err := t.fillPollResults(c, &poll); err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, poll)
}
